package main

import (
	"fmt"
	"sync"
	"time"
)

type Signal struct {
	ID           string
	CurrentState SignalState
	DurationSec  int
	Direction    Direction
	LastHeartAt  time.Time
	TrafficCount int
	mu           sync.Mutex
}

func (s *Signal) ChangeState(state SignalState) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	// 驗證狀態轉換是否合法
	if !isValidTransition(s.CurrentState, state) {
		return fmt.Errorf("invalid transition: %s -> %s", s.CurrentState, state)
	}
	s.CurrentState = state
	return nil
}

func (s *Signal) IsHealthy() bool {
	return time.Since(s.LastHeartAt) < 10*time.Second
}
