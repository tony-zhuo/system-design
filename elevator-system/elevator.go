package main

import "fmt"

// Elevator represents a single elevator car.
// It implements a LOOK algorithm (variant of SCAN):
//   - Serve all requests in the current direction first
//   - Reverse direction when no more requests ahead
type Elevator struct {
	ID           int
	CurrentFloor int
	State        ElevatorState
	Direction    Direction
	MinFloor     int
	MaxFloor     int

	// upStops and downStops track pending stops as boolean arrays.
	// Index maps directly to floor: index i represents floor (i + MinFloor).
	// upStops: floors to visit while going up
	// downStops: floors to visit while going down
	cabUpStops    []bool
	cabDownStops  []bool
	hallUpStops   []bool
	hallDownStops []bool

	// Cached boundaries of pending requests for O(1) direction checks.
	// When no requests: minRequest > maxRequest.
	minRequest int // lowest floor with a pending stop
	maxRequest int // highest floor with a pending stop

	// doorTimer counts down steps while the door is open.
	doorTimer int

	currentWeight int
	maxWeight     int
}

const doorOpenSteps = 2    // Number of steps the door stays open
const passengerWeight = 10 // Simulated weight per passenger boarding/alighting

// NewElevator creates an elevator starting at the given floor.
func NewElevator(id, minFloor, maxFloor int) *Elevator {
	n := maxFloor - minFloor + 1
	return &Elevator{
		ID:            id,
		CurrentFloor:  minFloor,
		State:         StateIdle,
		Direction:     DirIdle,
		MinFloor:      minFloor,
		MaxFloor:      maxFloor,
		hallUpStops:   make([]bool, n),
		hallDownStops: make([]bool, n),
		cabUpStops:    make([]bool, n),
		cabDownStops:  make([]bool, n),
		minRequest:    maxFloor + 1, // > maxRequest means empty
		maxRequest:    minFloor - 1,
		currentWeight: 0,
		maxWeight:     100,
	}
}

// idx converts a floor number to the array index.
func (e *Elevator) idx(floor int) int {
	return floor - e.MinFloor
}

// AddRequest adds a request to the elevator's stop sets using the LOOK strategy.
func (e *Elevator) AddRequest(r Request) {
	if r.Floor < e.MinFloor || r.Floor > e.MaxFloor {
		return
	}
	if r.Floor == e.CurrentFloor && (e.State == StateIdle || e.State == StateDoorOpen) {
		// Already at this floor and idle/door-open — open door, serve both directions.
		e.openDoor(DirIdle)
		return
	}

	i := e.idx(r.Floor)
	switch r.Type {
	case HallCall:
		// Hall call: place into the set matching the requested direction.
		if r.Direction == DirUp {
			e.hallUpStops[i] = true
		} else {
			e.hallDownStops[i] = true
		}
	case CabCall:
		// Cab call: place based on relative position and current direction.
		if r.Floor > e.CurrentFloor {
			e.cabUpStops[i] = true
		} else if r.Floor < e.CurrentFloor {
			e.cabDownStops[i] = true
		}
	}

	// Update cached boundaries.
	if r.Floor < e.minRequest {
		e.minRequest = r.Floor
	}
	if r.Floor > e.maxRequest {
		e.maxRequest = r.Floor
	}

	// If idle, start moving toward the request.
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

// Step advances the elevator by one time unit.
// Returns a human-readable description of what happened.
func (e *Elevator) Step() string {
	switch e.State {
	case StateDoorOpen:
		return e.stepDoorOpen()
	case StateMovingUp:
		return e.stepMove(DirUp)
	case StateMovingDown:
		return e.stepMove(DirDown)
	default: // StateIdle
		return e.stepIdle()
	}
}

func (e *Elevator) stepDoorOpen() string {
	e.doorTimer--
	if e.doorTimer > 0 {
		return fmt.Sprintf("Elevator %d: door open at floor %d (closing in %d)",
			e.ID, e.CurrentFloor, e.doorTimer)
	}
	// Door closes — decide next action.
	e.State = StateIdle
	e.pickDirection()
	return fmt.Sprintf("Elevator %d: door closed at floor %d, direction=%s",
		e.ID, e.CurrentFloor, e.Direction)
}

func (e *Elevator) stepMove(dir Direction) string {
	// Move one floor.
	if dir == DirUp {
		e.CurrentFloor++
	} else {
		e.CurrentFloor--
	}

	msg := fmt.Sprintf("Elevator %d: moved to floor %d", e.ID, e.CurrentFloor)

	// Check if we should stop here.
	if e.shouldStop(dir) {
		e.openDoor(dir)
		msg += " [STOP — door opening]"
	}
	return msg
}

func (e *Elevator) stepIdle() string {
	e.pickDirection()
	if e.State == StateIdle {
		return fmt.Sprintf("Elevator %d: idle at floor %d", e.ID, e.CurrentFloor)
	}
	// Started moving — delegate to move step next tick.
	return fmt.Sprintf("Elevator %d: idle at floor %d, starting %s",
		e.ID, e.CurrentFloor, e.Direction)
}

// shouldStop reports whether the elevator should stop at the current floor.
//   - Always stop if the current direction's stop set has this floor.
//   - At turnaround (no more stops ahead), also stop for the opposite direction's stop.
func (e *Elevator) shouldStop(dir Direction) bool {
	i := e.idx(e.CurrentFloor)

	if e.WeightSensor() {
		if dir == DirUp && !e.cabUpStops[i] && e.hallUpStops[i] {
			return false
		}
		if dir == DirDown && !e.cabDownStops[i] && e.hallDownStops[i] {
			return false
		}
	}

	if dir == DirUp {
		if e.cabUpStops[i] || e.hallUpStops[i] {
			return true
		}
		return !e.hasStopsAbove() && (e.cabDownStops[i] || e.hallDownStops[i])
	}
	if e.cabDownStops[i] || e.hallDownStops[i] {
		return true
	}
	return !e.hasStopsBelow() && (e.cabUpStops[i] || e.hallUpStops[i])
}

// openDoor transitions to door-open state and clears the served stops.
//
// Direction-aware clearing:
//   - DirUp:   clear upStops (intermediate stop serving up passengers)
//   - DirDown: clear downStops (intermediate stop serving down passengers)
//   - DirIdle: clear both (idle at this floor)
//
// Turnaround: when no more stops ahead in the current direction,
// also clear the opposite direction's stop (the elevator is reversing
// and won't revisit this floor in the new direction).
func (e *Elevator) openDoor(dir Direction) {
	e.State = StateDoorOpen
	e.doorTimer = doorOpenSteps
	i := e.idx(e.CurrentFloor)

	if dir == DirUp || dir == DirIdle {
		if e.cabUpStops[i] {
			e.cabUpStops[i] = false
			e.currentWeight -= passengerWeight
		}
		if e.hallUpStops[i] {
			e.hallUpStops[i] = false
			e.currentWeight += passengerWeight
		}
	}
	if dir == DirDown || dir == DirIdle {
		if e.cabDownStops[i] {
			e.cabDownStops[i] = false
			e.currentWeight -= passengerWeight
		}
		if e.hallDownStops[i] {
			e.hallDownStops[i] = false
			e.currentWeight += passengerWeight
		}
	}

	// Turnaround: also clear opposite direction stop.
	if dir == DirUp && !e.hasStopsAbove() {
		if e.cabDownStops[i] {
			e.cabDownStops[i] = false
			e.currentWeight -= passengerWeight
		}
		if e.hallDownStops[i] {
			e.hallDownStops[i] = false
			e.currentWeight += passengerWeight
		}
	}
	if dir == DirDown && !e.hasStopsBelow() {
		if e.cabUpStops[i] {
			e.cabUpStops[i] = false
			e.currentWeight -= passengerWeight
		}
		if e.hallUpStops[i] {
			e.hallUpStops[i] = false
			e.currentWeight += passengerWeight
		}
	}

	// Clamp weight to non-negative.
	if e.currentWeight < 0 {
		e.currentWeight = 0
	}

	// Recalculate bounds only if we just removed a boundary floor.
	if e.CurrentFloor == e.minRequest || e.CurrentFloor == e.maxRequest {
		e.recalcBounds()
	}
}

// recalcBounds rescans upStops and downStops to find new min/max.
func (e *Elevator) recalcBounds() {
	e.minRequest = e.MaxFloor + 1
	e.maxRequest = e.MinFloor - 1
	for i := range e.cabUpStops {
		if e.cabUpStops[i] || e.cabDownStops[i] || e.hallUpStops[i] || e.hallDownStops[i] {
			floor := i + e.MinFloor
			if floor < e.minRequest {
				e.minRequest = floor
			}
			if floor > e.maxRequest {
				e.maxRequest = floor
			}
		}
	}
}

// pickDirection decides the next direction based on pending requests (LOOK algorithm).
func (e *Elevator) pickDirection() {
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
		// Was idle — pick any direction with pending requests.
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
	// No pending requests.
	e.Direction = DirIdle
	e.State = StateIdle
}

// hasStopsAbove — O(1): compare current floor with cached maxRequest.
func (e *Elevator) hasStopsAbove() bool {
	return e.maxRequest > e.CurrentFloor
}

// hasStopsBelow — O(1): compare current floor with cached minRequest.
func (e *Elevator) hasStopsBelow() bool {
	return e.minRequest < e.CurrentFloor
}

// HasPendingRequests — O(1): check if bounds are valid.
func (e *Elevator) HasPendingRequests() bool {
	return e.minRequest <= e.maxRequest
}

// PendingCount returns the total number of pending stops.
func (e *Elevator) PendingCount() int {
	count := 0
	for i := range e.cabUpStops {
		if e.cabUpStops[i] || e.hallUpStops[i] {
			count++
		}
	}
	for i := range e.cabDownStops {
		if e.cabDownStops[i] || e.hallDownStops[i] {
			count++
		}
	}
	return count
}

// StopsCabSnapshot returns a copy of current stop sets for inspection.
func (e *Elevator) StopsCabSnapshot() (up []int, down []int) {
	for i, v := range e.cabUpStops {
		if v {
			up = append(up, i+e.MinFloor)
		}
	}
	for i, v := range e.cabDownStops {
		if v {
			down = append(down, i+e.MinFloor)
		}
	}
	return
}

// StopsHallSnapshot returns a copy of current stop sets for inspection.
func (e *Elevator) StopsHallSnapshot() (up []int, down []int) {
	for i, v := range e.hallUpStops {
		if v {
			up = append(up, i+e.MinFloor)
		}
	}
	for i, v := range e.hallDownStops {
		if v {
			down = append(down, i+e.MinFloor)
		}
	}
	return
}

// WeightSensor check is overweight
func (e *Elevator) WeightSensor() bool {
	return e.currentWeight >= e.maxWeight
}
