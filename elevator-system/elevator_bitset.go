package main

import (
	"fmt"

	"github.com/bits-and-blooms/bitset"
)

// BitsetElevator is an alternative Elevator implementation that uses
// github.com/bits-and-blooms/bitset for tracking stops.
//
// Compared to the bitmask (uint64) approach:
//   - No 64-floor limit — supports arbitrary building sizes
//   - Higher-level API: Set / Clear / Test / Count / NextSet / Any
//   - Slightly more overhead per operation due to abstraction
//
// Compared to []bool:
//   - Memory efficient: 1 bit per floor vs 1 byte per floor
//   - Count is O(n/64) via popcount instead of O(n) loop
//   - NextSet enables efficient hasStopsAbove without scanning every floor
type BitsetElevator struct {
	ID           int
	CurrentFloor int
	State        ElevatorState
	Direction    Direction
	MinFloor     int
	MaxFloor     int

	cabUpStops    *bitset.BitSet
	cabDownStops  *bitset.BitSet
	hallUpStops   *bitset.BitSet
	hallDownStops *bitset.BitSet

	doorTimer int

	currentWeight int
	maxWeight     int
}

// NewBitsetElevator creates an elevator using bitset stops.
func NewBitsetElevator(id, minFloor, maxFloor int) *BitsetElevator {
	n := uint(maxFloor - minFloor + 1)
	return &BitsetElevator{
		ID:            id,
		CurrentFloor:  minFloor,
		State:         StateIdle,
		Direction:     DirIdle,
		MinFloor:      minFloor,
		MaxFloor:      maxFloor,
		cabUpStops:    bitset.New(n),
		cabDownStops:  bitset.New(n),
		hallUpStops:   bitset.New(n),
		hallDownStops: bitset.New(n),
		maxWeight:     100,
	}
}

// idx converts a floor number to the bit position.
func (e *BitsetElevator) idx(floor int) uint {
	return uint(floor - e.MinFloor)
}

// --- Core elevator logic (same LOOK algorithm) ---

func (e *BitsetElevator) AddRequest(r Request) {
	if r.Floor < e.MinFloor || r.Floor > e.MaxFloor {
		return
	}
	if r.Floor == e.CurrentFloor && (e.State == StateIdle || e.State == StateDoorOpen) {
		e.openDoor(DirIdle)
		return
	}

	i := e.idx(r.Floor)
	switch r.Type {
	case HallCall:
		if r.Direction == DirUp {
			e.hallUpStops.Set(i)
		} else {
			e.hallDownStops.Set(i)
		}
	case CabCall:
		if r.Floor > e.CurrentFloor {
			e.cabUpStops.Set(i)
		} else if r.Floor < e.CurrentFloor {
			e.cabDownStops.Set(i)
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

func (e *BitsetElevator) Step() string {
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

func (e *BitsetElevator) stepDoorOpen() string {
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

func (e *BitsetElevator) stepMove(dir Direction) string {
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

func (e *BitsetElevator) stepIdle() string {
	e.pickDirection()
	if e.State == StateIdle {
		return fmt.Sprintf("Elevator %d: idle at floor %d", e.ID, e.CurrentFloor)
	}
	return fmt.Sprintf("Elevator %d: idle at floor %d, starting %s",
		e.ID, e.CurrentFloor, e.Direction)
}

func (e *BitsetElevator) shouldStop(dir Direction) bool {
	i := e.idx(e.CurrentFloor)

	if e.WeightSensor() {
		if dir == DirUp && !e.cabUpStops.Test(i) && e.hallUpStops.Test(i) {
			return false
		}
		if dir == DirDown && !e.cabDownStops.Test(i) && e.hallDownStops.Test(i) {
			return false
		}
	}

	if dir == DirUp {
		if e.cabUpStops.Test(i) || e.hallUpStops.Test(i) {
			return true
		}
		if !e.hasStopsAbove() && (e.cabDownStops.Test(i) || e.hallDownStops.Test(i)) {
			return true
		}
	} else {
		if e.cabDownStops.Test(i) || e.hallDownStops.Test(i) {
			return true
		}
		if !e.hasStopsBelow() && (e.cabUpStops.Test(i) || e.hallUpStops.Test(i)) {
			return true
		}
	}
	return false
}

func (e *BitsetElevator) openDoor(dir Direction) {
	e.State = StateDoorOpen
	e.doorTimer = doorOpenSteps
	i := e.idx(e.CurrentFloor)

	if dir == DirUp || dir == DirIdle {
		if e.cabUpStops.Test(i) {
			e.cabUpStops.Clear(i)
			e.currentWeight -= passengerWeight
		}
		if e.hallUpStops.Test(i) {
			e.hallUpStops.Clear(i)
			e.currentWeight += passengerWeight
		}
	}
	if dir == DirDown || dir == DirIdle {
		if e.cabDownStops.Test(i) {
			e.cabDownStops.Clear(i)
			e.currentWeight -= passengerWeight
		}
		if e.hallDownStops.Test(i) {
			e.hallDownStops.Clear(i)
			e.currentWeight += passengerWeight
		}
	}

	// Turnaround: also clear opposite direction stop.
	if dir == DirUp && !e.hasStopsAbove() {
		if e.cabDownStops.Test(i) {
			e.cabDownStops.Clear(i)
			e.currentWeight -= passengerWeight
		}
		if e.hallDownStops.Test(i) {
			e.hallDownStops.Clear(i)
			e.currentWeight += passengerWeight
		}
	}
	if dir == DirDown && !e.hasStopsBelow() {
		if e.cabUpStops.Test(i) {
			e.cabUpStops.Clear(i)
			e.currentWeight -= passengerWeight
		}
		if e.hallUpStops.Test(i) {
			e.hallUpStops.Clear(i)
			e.currentWeight += passengerWeight
		}
	}

	// Clamp weight to non-negative.
	if e.currentWeight < 0 {
		e.currentWeight = 0
	}
}

func (e *BitsetElevator) pickDirection() {
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

// hasStopsAbove uses NextSet to find the first set bit above the current floor.
func (e *BitsetElevator) hasStopsAbove() bool {
	start := e.idx(e.CurrentFloor) + 1
	n := uint(e.MaxFloor - e.MinFloor + 1)
	if next, ok := e.cabUpStops.NextSet(start); ok && next < n {
		return true
	}
	if next, ok := e.cabDownStops.NextSet(start); ok && next < n {
		return true
	}
	if next, ok := e.hallUpStops.NextSet(start); ok && next < n {
		return true
	}
	if next, ok := e.hallDownStops.NextSet(start); ok && next < n {
		return true
	}
	return false
}

// hasStopsBelow scans from bit 0 to idx-1 for any set bit.
func (e *BitsetElevator) hasStopsBelow() bool {
	cur := e.idx(e.CurrentFloor)
	if cur == 0 {
		return false
	}
	if next, ok := e.cabUpStops.NextSet(0); ok && next < cur {
		return true
	}
	if next, ok := e.cabDownStops.NextSet(0); ok && next < cur {
		return true
	}
	if next, ok := e.hallUpStops.NextSet(0); ok && next < cur {
		return true
	}
	if next, ok := e.hallDownStops.NextSet(0); ok && next < cur {
		return true
	}
	return false
}

// HasPendingRequests — checks if any bit is set.
func (e *BitsetElevator) HasPendingRequests() bool {
	return e.cabUpStops.Any() || e.cabDownStops.Any() ||
		e.hallUpStops.Any() || e.hallDownStops.Any()
}

// PendingCount returns the total number of pending stops.
func (e *BitsetElevator) PendingCount() int {
	return int(e.cabUpStops.Count() + e.cabDownStops.Count() +
		e.hallUpStops.Count() + e.hallDownStops.Count())
}

// StopsCabSnapshot returns the cab stop floors in each direction.
func (e *BitsetElevator) StopsCabSnapshot() (up []int, down []int) {
	n := uint(e.MaxFloor - e.MinFloor + 1)
	for i, ok := e.cabUpStops.NextSet(0); ok && i < n; i, ok = e.cabUpStops.NextSet(i + 1) {
		up = append(up, int(i)+e.MinFloor)
	}
	for i, ok := e.cabDownStops.NextSet(0); ok && i < n; i, ok = e.cabDownStops.NextSet(i + 1) {
		down = append(down, int(i)+e.MinFloor)
	}
	return
}

// StopsHallSnapshot returns the hall stop floors in each direction.
func (e *BitsetElevator) StopsHallSnapshot() (up []int, down []int) {
	n := uint(e.MaxFloor - e.MinFloor + 1)
	for i, ok := e.hallUpStops.NextSet(0); ok && i < n; i, ok = e.hallUpStops.NextSet(i + 1) {
		up = append(up, int(i)+e.MinFloor)
	}
	for i, ok := e.hallDownStops.NextSet(0); ok && i < n; i, ok = e.hallDownStops.NextSet(i + 1) {
		down = append(down, int(i)+e.MinFloor)
	}
	return
}

// WeightSensor check is overweight
func (e *BitsetElevator) WeightSensor() bool {
	return e.currentWeight >= e.maxWeight
}
