package main

import (
	"testing"
)

// runUntilIdle drives the elevator until it has no pending requests and is idle.
// Returns the sequence of floors where the elevator stopped (door opened).
func runUntilIdle(e *Elevator, maxSteps int) []int {
	var stops []int
	for i := 0; i < maxSteps; i++ {
		e.Step()
		if e.State == StateDoorOpen && e.doorTimer == doorOpenSteps {
			// Just opened the door at this floor.
			stops = append(stops, e.CurrentFloor)
		}
		if e.State == StateIdle && !e.HasPendingRequests() {
			break
		}
	}
	return stops
}

// --- Level 1: Basic single elevator ---

func TestElevator_BasicMoveUp(t *testing.T) {
	e := NewElevator(1, 1, 10)
	e.AddRequest(Request{Floor: 5, Type: CabCall})

	stops := runUntilIdle(e, 50)

	if len(stops) != 1 || stops[0] != 5 {
		t.Errorf("expected stops=[5], got %v", stops)
	}
	if e.CurrentFloor != 5 {
		t.Errorf("expected floor 5, got %d", e.CurrentFloor)
	}
}

func TestElevator_BasicMoveDown(t *testing.T) {
	e := NewElevator(1, 1, 10)
	e.CurrentFloor = 8
	e.AddRequest(Request{Floor: 3, Type: CabCall})

	stops := runUntilIdle(e, 50)

	if len(stops) != 1 || stops[0] != 3 {
		t.Errorf("expected stops=[3], got %v", stops)
	}
}

func TestElevator_DoorOpenDuration(t *testing.T) {
	e := NewElevator(1, 1, 10)
	e.AddRequest(Request{Floor: 2, Type: CabCall})

	// Step until door opens at floor 2.
	for i := 0; i < 50; i++ {
		e.Step()
		if e.State == StateDoorOpen && e.CurrentFloor == 2 {
			break
		}
	}

	if e.State != StateDoorOpen {
		t.Fatalf("expected DoorOpen state")
	}

	// Door should stay open for doorOpenSteps - 1 more steps.
	for i := 0; i < doorOpenSteps-1; i++ {
		e.Step()
		if i < doorOpenSteps-2 && e.State != StateDoorOpen {
			t.Errorf("door closed too early at step %d", i)
		}
	}
}

func TestElevator_AlreadyAtFloor(t *testing.T) {
	e := NewElevator(1, 1, 10)
	e.CurrentFloor = 5
	e.AddRequest(Request{Floor: 5, Type: CabCall})

	// Should immediately open door.
	if e.State != StateDoorOpen {
		t.Errorf("expected DoorOpen when already at requested floor, got %s", e.State)
	}
}

// --- Level 2: SCAN / LOOK algorithm ---

func TestElevator_SCANOrder_GoingUp(t *testing.T) {
	e := NewElevator(1, 1, 10)
	// Add requests: 7, 3, 5 — all cab calls.
	// Starting at floor 1, going up: should visit 3, 5, 7 in order.
	e.AddRequest(Request{Floor: 7, Type: CabCall})
	e.AddRequest(Request{Floor: 3, Type: CabCall})
	e.AddRequest(Request{Floor: 5, Type: CabCall})

	stops := runUntilIdle(e, 50)

	expected := []int{3, 5, 7}
	if !intSliceEqual(stops, expected) {
		t.Errorf("expected SCAN order %v, got %v", expected, stops)
	}
}

func TestElevator_SCANOrder_ReverseDirection(t *testing.T) {
	e := NewElevator(1, 1, 10)
	e.CurrentFloor = 5
	e.Direction = DirUp

	// Cab calls above and below.
	e.AddRequest(Request{Floor: 8, Type: CabCall})
	e.AddRequest(Request{Floor: 2, Type: CabCall})

	stops := runUntilIdle(e, 50)

	// Should go up to 8 first, then reverse to 2.
	expected := []int{8, 2}
	if !intSliceEqual(stops, expected) {
		t.Errorf("expected %v, got %v", expected, stops)
	}
}

func TestElevator_HallCall_DirectionFiltering(t *testing.T) {
	e := NewElevator(1, 1, 10)
	e.CurrentFloor = 1

	// Hall call at floor 5 going Up.
	e.AddRequest(Request{Floor: 5, Direction: DirUp, Type: HallCall})
	// Hall call at floor 3 going Down.
	e.AddRequest(Request{Floor: 3, Direction: DirDown, Type: HallCall})

	stops := runUntilIdle(e, 50)

	// Going up: stop at 5 (upStop). Then reverse: stop at 3 (downStop).
	expected := []int{5, 3}
	if !intSliceEqual(stops, expected) {
		t.Errorf("expected %v, got %v", expected, stops)
	}
}

func TestElevator_MixedRequests(t *testing.T) {
	e := NewElevator(1, 1, 10)
	e.CurrentFloor = 1

	// Hall call at floor 6 going Up, cab call to floor 4, cab call to floor 8.
	e.AddRequest(Request{Floor: 6, Direction: DirUp, Type: HallCall})
	e.AddRequest(Request{Floor: 4, Type: CabCall})
	e.AddRequest(Request{Floor: 8, Type: CabCall})

	stops := runUntilIdle(e, 50)

	// All going up: 4, 6, 8.
	expected := []int{4, 6, 8}
	if !intSliceEqual(stops, expected) {
		t.Errorf("expected %v, got %v", expected, stops)
	}
}

func TestElevator_NoRequests(t *testing.T) {
	e := NewElevator(1, 1, 10)
	msg := e.Step()

	if e.State != StateIdle {
		t.Errorf("expected Idle, got %s", e.State)
	}
	if msg == "" {
		t.Error("expected non-empty step message")
	}
}

func TestElevator_OutOfRangeRequest(t *testing.T) {
	e := NewElevator(1, 1, 10)
	e.AddRequest(Request{Floor: 15, Type: CabCall})

	if e.HasPendingRequests() {
		t.Error("should not accept out-of-range request")
	}
}

// --- Level 4: Overweight behavior ---

func TestElevator_Overweight_SkipsHallStop(t *testing.T) {
	e := NewElevator(1, 1, 10)
	e.currentWeight = e.maxWeight // overweight

	e.AddRequest(Request{Floor: 7, Type: CabCall})
	e.AddRequest(Request{Floor: 5, Direction: DirUp, Type: HallCall})

	stops := runUntilIdle(e, 100)

	// Normal order would be [5, 7]. Overweight skips hall at 5 on the way up;
	// after cab exit at 7 weight drops, so 5 is served on the return trip.
	expected := []int{7, 5}
	if !intSliceEqual(stops, expected) {
		t.Errorf("expected %v (hall at 5 skipped on first pass), got %v", expected, stops)
	}
}

func TestElevator_Overweight_StillServesCabStop(t *testing.T) {
	e := NewElevator(1, 1, 10)
	e.currentWeight = e.maxWeight

	e.AddRequest(Request{Floor: 3, Type: CabCall})
	e.AddRequest(Request{Floor: 6, Type: CabCall})

	stops := runUntilIdle(e, 100)

	expected := []int{3, 6}
	if !intSliceEqual(stops, expected) {
		t.Errorf("expected %v (cab stops unaffected), got %v", expected, stops)
	}
}

func TestElevator_Overweight_CabAndHallSameFloor(t *testing.T) {
	e := NewElevator(1, 1, 10)
	e.currentWeight = e.maxWeight

	e.AddRequest(Request{Floor: 5, Type: CabCall})
	e.AddRequest(Request{Floor: 5, Direction: DirUp, Type: HallCall})

	stops := runUntilIdle(e, 100)

	expected := []int{5}
	if !intSliceEqual(stops, expected) {
		t.Errorf("expected %v (cab request forces stop), got %v", expected, stops)
	}
}

func TestElevator_Overweight_WeightDrop_ResumesHallService(t *testing.T) {
	e := NewElevator(1, 1, 10)
	// Near max: after one cab passenger exits (-10), weight drops below max.
	e.currentWeight = e.maxWeight + passengerWeight - 1 // 109

	e.AddRequest(Request{Floor: 3, Type: CabCall})
	e.AddRequest(Request{Floor: 5, Direction: DirUp, Type: HallCall})
	e.AddRequest(Request{Floor: 7, Type: CabCall})

	stops := runUntilIdle(e, 100)

	// At floor 3: cab exit → weight drops to 99 (< 100) → no longer overweight.
	// At floor 5: hall stop now served normally.
	// At floor 7: cab stop.
	expected := []int{3, 5, 7}
	if !intSliceEqual(stops, expected) {
		t.Errorf("expected %v (weight drops at 3, resumes hall at 5), got %v", expected, stops)
	}
}

func intSliceEqual(a, b []int) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
