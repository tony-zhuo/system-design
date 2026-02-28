package main

import "fmt"

// Direction represents the movement direction of an elevator.
type Direction int

const (
	DirIdle Direction = iota
	DirUp
	DirDown
)

func (d Direction) String() string {
	switch d {
	case DirUp:
		return "Up"
	case DirDown:
		return "Down"
	default:
		return "Idle"
	}
}

// ElevatorState represents the current state of an elevator.
type ElevatorState int

const (
	StateIdle ElevatorState = iota
	StateMovingUp
	StateMovingDown
	StateDoorOpen
)

func (s ElevatorState) String() string {
	switch s {
	case StateMovingUp:
		return "MovingUp"
	case StateMovingDown:
		return "MovingDown"
	case StateDoorOpen:
		return "DoorOpen"
	default:
		return "Idle"
	}
}

// RequestType distinguishes between hall calls and cab calls.
type RequestType int

const (
	HallCall RequestType = iota // External button press (has direction)
	CabCall                     // Internal button press (destination only)
)

func (t RequestType) String() string {
	if t == HallCall {
		return "HallCall"
	}
	return "CabCall"
}

// Request represents an elevator request.
type Request struct {
	Floor     int
	Direction Direction   // Only meaningful for HallCall
	Type      RequestType
}

func (r Request) String() string {
	if r.Type == HallCall {
		return fmt.Sprintf("HallCall(floor=%d, dir=%s)", r.Floor, r.Direction)
	}
	return fmt.Sprintf("CabCall(floor=%d)", r.Floor)
}
