package main

import (
	"testing"
)

// runBitsetUntilIdle drives the bitset elevator until idle.
func runBitsetUntilIdle(e *BitsetElevator, maxSteps int) []int {
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

func TestBitsetElevator_BasicMoveUp(t *testing.T) {
	e := NewBitsetElevator(1, 1, 10)
	e.AddRequest(Request{Floor: 5, Type: CabCall})

	stops := runBitsetUntilIdle(e, 50)

	if len(stops) != 1 || stops[0] != 5 {
		t.Errorf("expected stops=[5], got %v", stops)
	}
}

func TestBitsetElevator_BasicMoveDown(t *testing.T) {
	e := NewBitsetElevator(1, 1, 10)
	e.CurrentFloor = 8
	e.AddRequest(Request{Floor: 3, Type: CabCall})

	stops := runBitsetUntilIdle(e, 50)

	if len(stops) != 1 || stops[0] != 3 {
		t.Errorf("expected stops=[3], got %v", stops)
	}
}

func TestBitsetElevator_AlreadyAtFloor(t *testing.T) {
	e := NewBitsetElevator(1, 1, 10)
	e.CurrentFloor = 5
	e.AddRequest(Request{Floor: 5, Type: CabCall})

	if e.State != StateDoorOpen {
		t.Errorf("expected DoorOpen when already at requested floor, got %s", e.State)
	}
}

// --- Level 2: SCAN / LOOK algorithm ---

func TestBitsetElevator_SCANOrder_GoingUp(t *testing.T) {
	e := NewBitsetElevator(1, 1, 10)
	e.AddRequest(Request{Floor: 7, Type: CabCall})
	e.AddRequest(Request{Floor: 3, Type: CabCall})
	e.AddRequest(Request{Floor: 5, Type: CabCall})

	stops := runBitsetUntilIdle(e, 50)

	expected := []int{3, 5, 7}
	if !intSliceEqual(stops, expected) {
		t.Errorf("expected SCAN order %v, got %v", expected, stops)
	}
}

func TestBitsetElevator_SCANOrder_ReverseDirection(t *testing.T) {
	e := NewBitsetElevator(1, 1, 10)
	e.CurrentFloor = 5
	e.Direction = DirUp

	e.AddRequest(Request{Floor: 8, Type: CabCall})
	e.AddRequest(Request{Floor: 2, Type: CabCall})

	stops := runBitsetUntilIdle(e, 50)

	expected := []int{8, 2}
	if !intSliceEqual(stops, expected) {
		t.Errorf("expected %v, got %v", expected, stops)
	}
}

func TestBitsetElevator_HallCall_DirectionFiltering(t *testing.T) {
	e := NewBitsetElevator(1, 1, 10)
	e.CurrentFloor = 1

	e.AddRequest(Request{Floor: 5, Direction: DirUp, Type: HallCall})
	e.AddRequest(Request{Floor: 3, Direction: DirDown, Type: HallCall})

	stops := runBitsetUntilIdle(e, 50)

	expected := []int{5, 3}
	if !intSliceEqual(stops, expected) {
		t.Errorf("expected %v, got %v", expected, stops)
	}
}

func TestBitsetElevator_MixedRequests(t *testing.T) {
	e := NewBitsetElevator(1, 1, 10)
	e.CurrentFloor = 1

	e.AddRequest(Request{Floor: 6, Direction: DirUp, Type: HallCall})
	e.AddRequest(Request{Floor: 4, Type: CabCall})
	e.AddRequest(Request{Floor: 8, Type: CabCall})

	stops := runBitsetUntilIdle(e, 50)

	expected := []int{4, 6, 8}
	if !intSliceEqual(stops, expected) {
		t.Errorf("expected %v, got %v", expected, stops)
	}
}

func TestBitsetElevator_OutOfRange(t *testing.T) {
	e := NewBitsetElevator(1, 1, 10)
	e.AddRequest(Request{Floor: 15, Type: CabCall})

	if e.HasPendingRequests() {
		t.Error("should not accept out-of-range request")
	}
}

func TestBitsetElevator_PendingCount(t *testing.T) {
	e := NewBitsetElevator(1, 1, 10)
	e.AddRequest(Request{Floor: 3, Type: CabCall})
	e.AddRequest(Request{Floor: 5, Type: CabCall})
	e.AddRequest(Request{Floor: 7, Type: CabCall})

	if got := e.PendingCount(); got != 3 {
		t.Errorf("expected PendingCount=3, got %d", got)
	}
}

// --- Verify all three implementations produce identical results ---

func TestBitsetElevator_MatchesOtherImpls(t *testing.T) {
	requests := []Request{
		{Floor: 7, Type: CabCall},
		{Floor: 3, Type: CabCall},
		{Floor: 5, Direction: DirUp, Type: HallCall},
		{Floor: 9, Type: CabCall},
	}

	boolElev := NewElevator(1, 1, 10)
	bitmaskElev := NewBitmaskElevator(1, 1, 10)
	bitsetElev := NewBitsetElevator(1, 1, 10)

	for _, r := range requests {
		boolElev.AddRequest(r)
		bitmaskElev.AddRequest(r)
		bitsetElev.AddRequest(r)
	}

	boolStops := runUntilIdle(boolElev, 50)
	bitmaskStops := runBitmaskUntilIdle(bitmaskElev, 50)
	bitsetStops := runBitsetUntilIdle(bitsetElev, 50)

	if !intSliceEqual(boolStops, bitsetStops) {
		t.Errorf("bitset diverged from []bool:\n  []bool:  %v\n  bitset:  %v", boolStops, bitsetStops)
	}
	if !intSliceEqual(bitmaskStops, bitsetStops) {
		t.Errorf("bitset diverged from bitmask:\n  bitmask: %v\n  bitset:  %v", bitmaskStops, bitsetStops)
	}
}
