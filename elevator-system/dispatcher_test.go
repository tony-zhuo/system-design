package main

import (
	"testing"
)

func TestDispatcher_ClosestElevatorSelected(t *testing.T) {
	d := NewDispatcher(3, 1, 10)
	// Place elevators at different floors.
	d.Elevators[0].CurrentFloor = 1
	d.Elevators[1].CurrentFloor = 5
	d.Elevators[2].CurrentFloor = 9

	// Hall call at floor 6 going up — elevator 2 (floor 5) should be selected.
	chosen := d.Dispatch(Request{Floor: 6, Direction: DirUp, Type: HallCall})

	if chosen.ID != 2 {
		t.Errorf("expected elevator 2, got elevator %d", chosen.ID)
	}
}

func TestDispatcher_SameDirectionPreferred(t *testing.T) {
	d := NewDispatcher(2, 1, 10)

	// Elevator 1 at floor 3 moving up.
	d.Elevators[0].CurrentFloor = 3
	d.Elevators[0].State = StateMovingUp
	d.Elevators[0].Direction = DirUp

	// Elevator 2 at floor 4 moving down.
	d.Elevators[1].CurrentFloor = 4
	d.Elevators[1].State = StateMovingDown
	d.Elevators[1].Direction = DirDown

	// Hall call at floor 6 going up — elevator 1 (moving up) should be preferred
	// even though elevator 2 is slightly closer.
	chosen := d.Dispatch(Request{Floor: 6, Direction: DirUp, Type: HallCall})

	if chosen.ID != 1 {
		t.Errorf("expected elevator 1 (same direction), got elevator %d", chosen.ID)
	}
}

func TestDispatcher_IdleElevatorPreferred(t *testing.T) {
	d := NewDispatcher(2, 1, 10)

	// Elevator 1 at floor 5, idle.
	d.Elevators[0].CurrentFloor = 5

	// Elevator 2 at floor 5, moving up with pending requests.
	d.Elevators[1].CurrentFloor = 5
	d.Elevators[1].State = StateMovingUp
	d.Elevators[1].Direction = DirUp
	d.Elevators[1].AddRequest(Request{Floor: 8, Type: CabCall})
	d.Elevators[1].AddRequest(Request{Floor: 9, Type: CabCall})

	// Hall call at floor 3 going down.
	chosen := d.Dispatch(Request{Floor: 3, Direction: DirDown, Type: HallCall})

	if chosen.ID != 1 {
		t.Errorf("expected idle elevator 1, got elevator %d", chosen.ID)
	}
}

func TestDispatcher_StepAll(t *testing.T) {
	d := NewDispatcher(2, 1, 10)

	d.Dispatch(Request{Floor: 5, Direction: DirUp, Type: HallCall})
	d.Dispatch(Request{Floor: 3, Direction: DirDown, Type: HallCall})

	msgs := d.StepAll()
	if len(msgs) != 2 {
		t.Errorf("expected 2 messages, got %d", len(msgs))
	}

	for _, m := range msgs {
		if m == "" {
			t.Error("expected non-empty step message")
		}
	}
}

func TestDispatcher_MultipleRequests_Distribution(t *testing.T) {
	d := NewDispatcher(2, 1, 10)

	// Send two hall calls to opposite ends — should be distributed.
	e1 := d.Dispatch(Request{Floor: 2, Direction: DirUp, Type: HallCall})
	e2 := d.Dispatch(Request{Floor: 9, Direction: DirDown, Type: HallCall})

	if e1.ID == e2.ID {
		t.Errorf("expected different elevators for opposite requests, both got %d", e1.ID)
	}
}

func TestDispatcher_AllIdle(t *testing.T) {
	d := NewDispatcher(2, 1, 10)

	if !d.AllIdle() {
		t.Error("expected all elevators to be idle initially")
	}

	d.Dispatch(Request{Floor: 5, Direction: DirUp, Type: HallCall})
	if d.AllIdle() {
		t.Error("expected not all idle after dispatching a request")
	}
}
