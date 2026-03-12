package main

type Controllable interface {
	ChangeState(state SignalState) error
}

type HealthChecker interface {
	IsHealthy() bool
}

type SchedulePolicy interface {
	NextDirection(intersection *Intersection) Direction
}
