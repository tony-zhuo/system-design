package main

import (
	"fmt"
	"math/bits"
)

// BitmaskElevator is an alternative Elevator implementation that uses uint64
// bitmasks instead of []bool for tracking stops.
//
// Each bit in the uint64 represents a floor: bit i = floor (i + MinFloor).
// This limits the building to 64 floors, which covers most real-world cases.
//
// Advantages over []bool:
//   - hasStopsAbove / hasStopsBelow become O(1) bit masking instead of O(n) loop
//   - HasPendingRequests is O(1): just check bits != 0
//   - PendingCount is O(1): bits.OnesCount64
//   - Memory: 2 × 8 bytes = 16 bytes vs 2 × N bytes for []bool
type BitmaskElevator struct {
	ID           int
	CurrentFloor int
	State        ElevatorState
	Direction    Direction
	MinFloor     int
	MaxFloor     int

	upStops   uint64 // bitmask: bit i = floor (i + MinFloor) needs stop going up
	downStops uint64 // bitmask: bit i = floor (i + MinFloor) needs stop going down

	doorTimer int
}

const bitmaskMaxFloors = 64

// NewBitmaskElevator creates an elevator using bitmask stops.
// Panics if maxFloor - minFloor >= 64.
func NewBitmaskElevator(id, minFloor, maxFloor int) *BitmaskElevator {
	if maxFloor-minFloor+1 > bitmaskMaxFloors {
		panic(fmt.Sprintf("bitmask elevator supports at most %d floors", bitmaskMaxFloors))
	}
	return &BitmaskElevator{
		ID:           id,
		CurrentFloor: minFloor,
		State:        StateIdle,
		Direction:    DirIdle,
		MinFloor:     minFloor,
		MaxFloor:     maxFloor,
	}
}

// --- Bit manipulation helpers ---

// idx converts a floor number to the bit position.
func (e *BitmaskElevator) idx(floor int) uint {
	return uint(floor - e.MinFloor)
}

// set sets bit for the given floor.
func set(bitmask *uint64, bit uint) { *bitmask |= 1 << bit }

// clear clears bit for the given floor.
func clear(bitmask *uint64, bit uint) { *bitmask &^= 1 << bit }

// has checks if the bit for the given floor is set.
func has(bitmask uint64, bit uint) bool { return bitmask&(1<<bit) != 0 }

// aboveMask returns a mask with all bits above position `bit` set.
//
//	bit=2 → 0b...11111000
func aboveMask(bit uint) uint64 {
	if bit >= 63 {
		return 0
	}
	return ^uint64(0) << (bit + 1)
}

// belowMask returns a mask with all bits below position `bit` set.
//
//	bit=3 → 0b...00000111
func belowMask(bit uint) uint64 {
	if bit == 0 {
		return 0
	}
	return (1 << bit) - 1
}

// --- Core elevator logic (same LOOK algorithm, different data structure) ---

func (e *BitmaskElevator) AddRequest(r Request) {
	if r.Floor < e.MinFloor || r.Floor > e.MaxFloor {
		return
	}
	if r.Floor == e.CurrentFloor && (e.State == StateIdle || e.State == StateDoorOpen) {
		e.openDoor(DirIdle)
		return
	}

	bit := e.idx(r.Floor)
	switch r.Type {
	case HallCall:
		if r.Direction == DirUp {
			set(&e.upStops, bit)
		} else {
			set(&e.downStops, bit)
		}
	case CabCall:
		if r.Floor > e.CurrentFloor {
			set(&e.upStops, bit)
		} else if r.Floor < e.CurrentFloor {
			set(&e.downStops, bit)
		}
	}

	if e.State == StateIdle {
		if r.Floor > e.CurrentFloor {
			e.Direction = DirUp
			e.State = StateMovingUp
		} else if r.Floor < e.CurrentFloor {
			e.Direction = DirDown
			e.State = StateMovingDown
		}
	}
}

func (e *BitmaskElevator) Step() string {
	switch e.State {
	case StateDoorOpen:
		return e.stepDoorOpen()
	case StateMovingUp:
		return e.stepMove(DirUp)
	case StateMovingDown:
		return e.stepMove(DirDown)
	default:
		return e.stepIdle()
	}
}

func (e *BitmaskElevator) stepDoorOpen() string {
	e.doorTimer--
	if e.doorTimer > 0 {
		return fmt.Sprintf("Elevator %d: door open at floor %d (closing in %d)",
			e.ID, e.CurrentFloor, e.doorTimer)
	}
	e.State = StateIdle
	e.pickDirection()
	return fmt.Sprintf("Elevator %d: door closed at floor %d, direction=%s",
		e.ID, e.CurrentFloor, e.Direction)
}

func (e *BitmaskElevator) stepMove(dir Direction) string {
	if dir == DirUp {
		e.CurrentFloor++
	} else {
		e.CurrentFloor--
	}

	msg := fmt.Sprintf("Elevator %d: moved to floor %d", e.ID, e.CurrentFloor)
	if e.shouldStop(dir) {
		e.openDoor(dir)
		msg += " [STOP — door opening]"
	}
	return msg
}

func (e *BitmaskElevator) stepIdle() string {
	e.pickDirection()
	if e.State == StateIdle {
		return fmt.Sprintf("Elevator %d: idle at floor %d", e.ID, e.CurrentFloor)
	}
	return fmt.Sprintf("Elevator %d: idle at floor %d, starting %s",
		e.ID, e.CurrentFloor, e.Direction)
}

// shouldStop — O(1) with bitmask operations.
func (e *BitmaskElevator) shouldStop(dir Direction) bool {
	bit := e.idx(e.CurrentFloor)
	if dir == DirUp {
		if has(e.upStops, bit) {
			return true
		}
		if !e.hasStopsAbove() && has(e.downStops, bit) {
			return true
		}
	} else {
		if has(e.downStops, bit) {
			return true
		}
		if !e.hasStopsBelow() && has(e.upStops, bit) {
			return true
		}
	}
	return false
}

func (e *BitmaskElevator) openDoor(dir Direction) {
	e.State = StateDoorOpen
	e.doorTimer = doorOpenSteps
	bit := e.idx(e.CurrentFloor)

	if dir == DirUp || dir == DirIdle {
		clear(&e.upStops, bit)
	}
	if dir == DirDown || dir == DirIdle {
		clear(&e.downStops, bit)
	}

	// Turnaround: also clear opposite direction stop.
	if dir == DirUp && !e.hasStopsAbove() {
		clear(&e.downStops, bit)
	}
	if dir == DirDown && !e.hasStopsBelow() {
		clear(&e.upStops, bit)
	}
}

func (e *BitmaskElevator) pickDirection() {
	switch e.Direction {
	case DirUp:
		if e.hasStopsAbove() {
			e.State = StateMovingUp
			return
		}
		if e.hasStopsBelow() {
			e.Direction = DirDown
			e.State = StateMovingDown
			return
		}
	case DirDown:
		if e.hasStopsBelow() {
			e.State = StateMovingDown
			return
		}
		if e.hasStopsAbove() {
			e.Direction = DirUp
			e.State = StateMovingUp
			return
		}
	default:
		if e.hasStopsAbove() {
			e.Direction = DirUp
			e.State = StateMovingUp
			return
		}
		if e.hasStopsBelow() {
			e.Direction = DirDown
			e.State = StateMovingDown
			return
		}
	}
	e.Direction = DirIdle
	e.State = StateIdle
}

// hasStopsAbove — O(1): mask off bits above current floor, check != 0.
func (e *BitmaskElevator) hasStopsAbove() bool {
	mask := aboveMask(e.idx(e.CurrentFloor))
	return (e.upStops|e.downStops)&mask != 0
}

// hasStopsBelow — O(1): mask off bits below current floor, check != 0.
func (e *BitmaskElevator) hasStopsBelow() bool {
	mask := belowMask(e.idx(e.CurrentFloor))
	return (e.upStops|e.downStops)&mask != 0
}

// HasPendingRequests — O(1): just check if any bit is set.
func (e *BitmaskElevator) HasPendingRequests() bool {
	return (e.upStops | e.downStops) != 0
}

// PendingCount — O(1): popcount via bits.OnesCount64.
func (e *BitmaskElevator) PendingCount() int {
	return bits.OnesCount64(e.upStops) + bits.OnesCount64(e.downStops)
}

// StopsSnapshot returns the floors in each stop set.
func (e *BitmaskElevator) StopsSnapshot() (up []int, down []int) {
	for b := e.upStops; b != 0; {
		i := bits.TrailingZeros64(b) // lowest set bit
		up = append(up, i+e.MinFloor)
		b &= b - 1 // clear lowest set bit
	}
	for b := e.downStops; b != 0; {
		i := bits.TrailingZeros64(b)
		down = append(down, i+e.MinFloor)
		b &= b - 1
	}
	return
}
