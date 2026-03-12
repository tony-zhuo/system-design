package main

import "sync"

type FixedPolicy struct {
	directions []Direction
	current    int
	mu         sync.Mutex
}

func (f *FixedPolicy) NextDirection(intersection *Intersection) Direction {
	f.mu.Lock()
	defer f.mu.Unlock()

	dir := f.directions[f.current]
	f.current = (f.current + 1) % len(f.directions)
	return dir
}

type DynamicPolicy struct {
	mu sync.RWMutex
}

func (d *DynamicPolicy) NextDirection(intersection *Intersection) Direction {
	d.mu.RLock()
	defer d.mu.RUnlock()

	// 找出車流量最高的方向
	var bestDir Direction
	maxCount := -1

	for dir, signal := range intersection.Signals {
		if signal.TrafficCount > maxCount {
			maxCount = signal.TrafficCount
			bestDir = dir
		}
	}
	return bestDir
}
