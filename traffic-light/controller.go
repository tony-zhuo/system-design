package main

import (
	"sync"
	"time"
)

type Controller struct {
	// 需要哪些欄位？
	Intersections map[string]*Intersection // id: Intersection
	Policy        SchedulePolicy
	mu            sync.RWMutex
}

func (c *Controller) Run() {
	for {
		c.mu.RLock()
		for _, intersection := range c.Intersections {
			switch intersection.CurrentMode {
			case ModeEmergency, ModeManual:
				continue // 緊急/人工模式，跳過自動調度
			default:
				dir := c.Policy.NextDirection(intersection)
				intersection.ActivateDirection(dir)
			}
		}
		c.mu.RUnlock()
		time.Sleep(5 * time.Second)
	}
}

func (c *Controller) SetEmergencyMode(intersectionID string, dir Direction) {
	c.mu.Lock()
	defer c.mu.Unlock()

	intersection := c.Intersections[intersectionID]
	intersection.CurrentMode = ModeEmergency
	intersection.ActivateDirection(dir)
}

func (c *Controller) ClearEmergencyMode(intersectionID string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	intersection := c.Intersections[intersectionID]
	intersection.CurrentMode = ModeNormal
	// Controller 的 Run() loop 會自動接管
}
