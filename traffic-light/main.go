package main

/*
┌─────────────────────────────────┐
│           Controller            │
│                                 │
│  policy: SchedulePolicy         │
│  intersections: []*Intersection │
│                                 │
│  Run() → 問 Policy → 叫         │
│          Intersection 切換      │
└─────────────────────────────────┘
           │              │
           ▼              ▼
    SchedulePolicy    Intersection
    ┌────────────┐    ┌───────────────┐
    │FixedPolicy │    │ signals:      │
    │DynamicPolicy│   │  North:Signal │
    └────────────┘    │  South:Signal │
                      │  East: Signal │
                      │  West: Signal │
                      └───────────────┘
                               │
                               ▼
                           Signal
                      ┌───────────────┐
                      │ CurrentState  │
                      │ Direction     │
                      │ LastHeartAt   │
                      │ mu sync.Mutex │
                      └───────────────┘
*/

func main() {

}
