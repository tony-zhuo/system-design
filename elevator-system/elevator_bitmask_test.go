package main

import (
	"testing"
)

// runBitmaskUntilIdle drives the bitmask elevator until idle.
func runBitmaskUntilIdle(e *BitmaskElevator, maxSteps int) []int {
	var stops []int
	for range maxSteps {
		e.Step()
		if e.State == StateDoorOpen && e.doorTimer == doorOpenSteps {
			stops = append(stops, e.CurrentFloor)
		}
		if e.State == StateIdle && !e.HasPendingRequests() {
			break
		}
	}
	return stops
}

// --- Level 1: Basic single elevator ---

func TestBitmaskElevator_BasicMoveUp(t *testing.T) {
	e := NewBitmaskElevator(1, 1, 10)
	e.AddRequest(Request{Floor: 5, Type: CabCall})

	stops := runBitmaskUntilIdle(e, 50)

	if len(stops) != 1 || stops[0] != 5 {
		t.Errorf("expected stops=[5], got %v", stops)
	}
}

func TestBitmaskElevator_BasicMoveDown(t *testing.T) {
	e := NewBitmaskElevator(1, 1, 10)
	e.CurrentFloor = 8
	e.AddRequest(Request{Floor: 3, Type: CabCall})

	stops := runBitmaskUntilIdle(e, 50)

	if len(stops) != 1 || stops[0] != 3 {
		t.Errorf("expected stops=[3], got %v", stops)
	}
}

func TestBitmaskElevator_AlreadyAtFloor(t *testing.T) {
	e := NewBitmaskElevator(1, 1, 10)
	e.CurrentFloor = 5
	e.AddRequest(Request{Floor: 5, Type: CabCall})

	if e.State != StateDoorOpen {
		t.Errorf("expected DoorOpen when already at requested floor, got %s", e.State)
	}
}

// --- Level 2: SCAN / LOOK algorithm ---

func TestBitmaskElevator_SCANOrder_GoingUp(t *testing.T) {
	e := NewBitmaskElevator(1, 1, 10)
	e.AddRequest(Request{Floor: 7, Type: CabCall})
	e.AddRequest(Request{Floor: 3, Type: CabCall})
	e.AddRequest(Request{Floor: 5, Type: CabCall})

	stops := runBitmaskUntilIdle(e, 50)

	expected := []int{3, 5, 7}
	if !intSliceEqual(stops, expected) {
		t.Errorf("expected SCAN order %v, got %v", expected, stops)
	}
}

func TestBitmaskElevator_SCANOrder_ReverseDirection(t *testing.T) {
	e := NewBitmaskElevator(1, 1, 10)
	e.CurrentFloor = 5
	e.Direction = DirUp

	e.AddRequest(Request{Floor: 8, Type: CabCall})
	e.AddRequest(Request{Floor: 2, Type: CabCall})

	stops := runBitmaskUntilIdle(e, 50)

	expected := []int{8, 2}
	if !intSliceEqual(stops, expected) {
		t.Errorf("expected %v, got %v", expected, stops)
	}
}

func TestBitmaskElevator_HallCall_DirectionFiltering(t *testing.T) {
	e := NewBitmaskElevator(1, 1, 10)
	e.CurrentFloor = 1

	e.AddRequest(Request{Floor: 5, Direction: DirUp, Type: HallCall})
	e.AddRequest(Request{Floor: 3, Direction: DirDown, Type: HallCall})

	stops := runBitmaskUntilIdle(e, 50)

	expected := []int{5, 3}
	if !intSliceEqual(stops, expected) {
		t.Errorf("expected %v, got %v", expected, stops)
	}
}

func TestBitmaskElevator_MixedRequests(t *testing.T) {
	e := NewBitmaskElevator(1, 1, 10)
	e.CurrentFloor = 1

	e.AddRequest(Request{Floor: 6, Direction: DirUp, Type: HallCall})
	e.AddRequest(Request{Floor: 4, Type: CabCall})
	e.AddRequest(Request{Floor: 8, Type: CabCall})

	stops := runBitmaskUntilIdle(e, 50)

	expected := []int{4, 6, 8}
	if !intSliceEqual(stops, expected) {
		t.Errorf("expected %v, got %v", expected, stops)
	}
}

func TestBitmaskElevator_OutOfRange(t *testing.T) {
	e := NewBitmaskElevator(1, 1, 10)
	e.AddRequest(Request{Floor: 15, Type: CabCall})

	if e.HasPendingRequests() {
		t.Error("should not accept out-of-range request")
	}
}

func TestBitmaskElevator_PendingCount(t *testing.T) {
	e := NewBitmaskElevator(1, 1, 10)
	e.AddRequest(Request{Floor: 3, Type: CabCall})
	e.AddRequest(Request{Floor: 5, Type: CabCall})
	e.AddRequest(Request{Floor: 7, Type: CabCall})

	if got := e.PendingCount(); got != 3 {
		t.Errorf("expected PendingCount=3, got %d", got)
	}
}

// --- Verify both implementations produce identical results ---

func TestBitmaskElevator_MatchesBoolArray(t *testing.T) {
	requests := []Request{
		{Floor: 7, Type: CabCall},
		{Floor: 3, Type: CabCall},
		{Floor: 5, Direction: DirUp, Type: HallCall},
		{Floor: 9, Type: CabCall},
	}

	boolElev := NewElevator(1, 1, 10)
	bitmaskElev := NewBitmaskElevator(1, 1, 10)

	for _, r := range requests {
		boolElev.AddRequest(r)
		bitmaskElev.AddRequest(r)
	}

	boolStops := runUntilIdle(boolElev, 50)
	bitmaskStops := runBitmaskUntilIdle(bitmaskElev, 50)

	if !intSliceEqual(boolStops, bitmaskStops) {
		t.Errorf("implementations diverged:\n  []bool:   %v\n  bitmask:  %v", boolStops, bitmaskStops)
	}
}
