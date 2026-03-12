package main

import "sync"

type Intersection struct {
	ID          string
	Signals     map[Direction]*Signal
	CurrentMode Mode
	mu          sync.RWMutex
}

func (i *Intersection) IsHealthy() bool {
	for _, signal := range i.Signals {
		if !signal.IsHealthy() {
			return false
		}
	}
	return true
}

func (i *Intersection) ActivateDirection(dir Direction) error {
	for direction, signal := range i.Signals {
		if direction == dir {
			signal.ChangeState(Green)
		} else {
			signal.ChangeState(Red)
		}
	}
	return nil
}
