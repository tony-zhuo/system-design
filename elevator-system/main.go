package main

import "fmt"

func main() {
	fmt.Println("========================================")
	fmt.Println(" Elevator System Design — Demo")
	fmt.Println("========================================")

	demoLevel1()
	demoLevel2()
	demoLevel3()
}

func demoLevel1() {
	fmt.Println("\n--- Level 1: Single Elevator (Basic) ---")
	fmt.Println("Scenario: Elevator at floor 1, cab call to floor 5")
	fmt.Println()

	e := NewElevator(1, 1, 10)
	e.AddRequest(Request{Floor: 5, Type: CabCall})

	for i := 1; i <= 20; i++ {
		msg := e.Step()
		fmt.Printf("  Step %2d: %s\n", i, msg)
		if e.State == StateIdle && !e.HasPendingRequests() {
			break
		}
	}
}

func demoLevel2() {
	fmt.Println("\n--- Level 2: SCAN Scheduling (LOOK Algorithm) ---")
	fmt.Println("Scenario: Elevator at floor 1")
	fmt.Println("  Requests: CabCall(7), CabCall(3), HallCall(5,Up), CabCall(9)")
	fmt.Println("  Expected SCAN order going up: 3 → 5 → 7 → 9")
	fmt.Println()

	e := NewElevator(1, 1, 10)
	e.AddRequest(Request{Floor: 7, Type: CabCall})
	e.AddRequest(Request{Floor: 3, Type: CabCall})
	e.AddRequest(Request{Floor: 5, Direction: DirUp, Type: HallCall})
	e.AddRequest(Request{Floor: 9, Type: CabCall})

	for i := 1; i <= 50; i++ {
		msg := e.Step()
		fmt.Printf("  Step %2d: %s\n", i, msg)
		if e.State == StateIdle && !e.HasPendingRequests() {
			break
		}
	}

	fmt.Println("\nScenario: Elevator at floor 5, direction=Up")
	fmt.Println("  Requests: CabCall(8), CabCall(2), HallCall(3,Down)")
	fmt.Println("  Expected: Up to 8, then reverse down to 3, 2")
	fmt.Println()

	e2 := NewElevator(1, 1, 10)
	e2.CurrentFloor = 5
	e2.Direction = DirUp
	e2.AddRequest(Request{Floor: 8, Type: CabCall})
	e2.AddRequest(Request{Floor: 2, Type: CabCall})
	e2.AddRequest(Request{Floor: 3, Direction: DirDown, Type: HallCall})

	for i := 1; i <= 50; i++ {
		msg := e2.Step()
		fmt.Printf("  Step %2d: %s\n", i, msg)
		if e2.State == StateIdle && !e2.HasPendingRequests() {
			break
		}
	}
}

func demoLevel3() {
	fmt.Println("\n--- Level 3: Multi-Elevator Dispatch ---")
	fmt.Println("Scenario: 3 elevators, 10 floors")
	fmt.Println("  E1 at floor 1, E2 at floor 5, E3 at floor 9")
	fmt.Println()

	d := NewDispatcher(3, 1, 10)
	d.Elevators[0].CurrentFloor = 1
	d.Elevators[1].CurrentFloor = 5
	d.Elevators[2].CurrentFloor = 9

	// Dispatch several hall calls.
	requests := []Request{
		{Floor: 3, Direction: DirUp, Type: HallCall},
		{Floor: 7, Direction: DirDown, Type: HallCall},
		{Floor: 2, Direction: DirUp, Type: HallCall},
		{Floor: 8, Direction: DirUp, Type: HallCall},
	}

	for _, r := range requests {
		chosen := d.Dispatch(r)
		fmt.Printf("  Dispatched %s → Elevator %d\n", r, chosen.ID)
	}

	fmt.Println("\nRunning simulation:")
	fmt.Println(d.Status())

	for i := 1; i <= 30; i++ {
		msgs := d.StepAll()
		anyActive := false
		for _, m := range msgs {
			fmt.Printf("  Step %2d: %s\n", i, m)
		}
		for _, e := range d.Elevators {
			if e.State != StateIdle || e.HasPendingRequests() {
				anyActive = true
			}
		}
		if !anyActive {
			fmt.Println("\n  All elevators idle.")
			break
		}
	}
}
