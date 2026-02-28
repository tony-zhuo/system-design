package main

import (
	"fmt"
	"math"
)

// Dispatcher manages multiple elevators and assigns hall calls to the best one.
type Dispatcher struct {
	Elevators []*Elevator
	MinFloor  int
	MaxFloor  int
}

// NewDispatcher creates a dispatcher with n elevators.
func NewDispatcher(n, minFloor, maxFloor int) *Dispatcher {
	elevators := make([]*Elevator, n)
	for i := range n {
		elevators[i] = NewElevator(i+1, minFloor, maxFloor)
	}
	return &Dispatcher{
		Elevators: elevators,
		MinFloor:  minFloor,
		MaxFloor:  maxFloor,
	}
}

// Dispatch assigns a hall call to the best elevator using a cost function.
// The cost considers:
//   - Distance from the elevator to the request floor
//   - Direction alignment bonus (same direction = lower cost)
//   - Current load (number of pending requests)
func (d *Dispatcher) Dispatch(r Request) *Elevator {
	if len(d.Elevators) == 0 {
		return nil
	}

	var best *Elevator
	bestCost := math.MaxFloat64

	for _, e := range d.Elevators {
		cost := d.cost(e, r)
		if cost < bestCost {
			bestCost = cost
			best = e
		}
	}

	if best != nil {
		best.AddRequest(r)
	}
	return best
}

// cost calculates the cost for an elevator to serve a request.
//
// Cost formula:
//
//	base = |currentFloor - requestFloor|
//	if elevator is idle: cost = base
//	if elevator is moving toward the request and same direction: cost = base
//	if elevator is moving toward but opposite direction: cost = base + N/2
//	if elevator is moving away: cost = distance_to_end + end_to_request
//
// A small penalty is added for each pending request to prefer less-loaded elevators.
func (d *Dispatcher) cost(e *Elevator, r Request) float64 {
	distance := abs(e.CurrentFloor - r.Floor)

	// Idle elevator: pure distance.
	if e.State == StateIdle || e.Direction == DirIdle {
		return float64(distance) + 0.5*float64(e.PendingCount())
	}

	movingToward := (e.Direction == DirUp && r.Floor >= e.CurrentFloor) ||
		(e.Direction == DirDown && r.Floor <= e.CurrentFloor)

	if movingToward {
		sameDir := r.Type == CabCall || r.Direction == e.Direction
		if sameDir {
			// Best case: on the way and same direction.
			return float64(distance) + 0.5*float64(e.PendingCount())
		}
		// On the way but opposite direction — will pass through but won't pick up.
		// Needs to go to end first, then come back.
		span := float64(d.MaxFloor - d.MinFloor)
		return float64(distance) + span/2 + 0.5*float64(e.PendingCount())
	}

	// Moving away: must go to end, reverse, then reach the floor.
	var detour int
	if e.Direction == DirUp {
		detour = (d.MaxFloor - e.CurrentFloor) + (d.MaxFloor - r.Floor)
	} else {
		detour = (e.CurrentFloor - d.MinFloor) + (r.Floor - d.MinFloor)
	}
	return float64(detour) + 0.5*float64(e.PendingCount())
}

// StepAll advances all elevators by one time unit.
// Returns descriptions of each elevator's action.
func (d *Dispatcher) StepAll() []string {
	msgs := make([]string, len(d.Elevators))
	for i, e := range d.Elevators {
		msgs[i] = e.Step()
	}
	return msgs
}

// AllIdle returns true if every elevator is idle with no pending requests.
func (d *Dispatcher) AllIdle() bool {
	for _, e := range d.Elevators {
		if e.State != StateIdle || e.HasPendingRequests() {
			return false
		}
	}
	return true
}

// Status returns a summary string of all elevators.
func (d *Dispatcher) Status() string {
	s := ""
	for _, e := range d.Elevators {
		s += fmt.Sprintf("  [E%d] floor=%d state=%s dir=%s pending=%d\n",
			e.ID, e.CurrentFloor, e.State, e.Direction, e.PendingCount())
	}
	return s
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}
