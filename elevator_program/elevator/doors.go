package elevator

import (
	"elevator_program/elevio"
	"fmt"
	"time"
)

type DoorState int

const (
	DS_Closed DoorState = iota
	DS_Closeing
	DS_Open
	DS_Opening
	DS_Obstruction
	DS_Error
)

func (e *Elevator) updateDoorState() {
	switch e.doorState {
	case DS_Closed:
		return

	case DS_Opening:
		elevio.SetDoorOpenLamp(true)

		if e.atTargetFloor() {
			e.clearCurrentFloor()
		}

		e.doorState = DS_Open
		e.doorStartTimer = time.Time{}

	case DS_Open:
		if e.obstruction {
			e.doorState = DS_Obstruction
			e.doorStartTimer = time.Time{} // reset timer
			break
		}

		if e.doorStartTimer.IsZero() {
			e.doorStartTimer = time.Now()
		}

		if time.Since(e.doorStartTimer) >= 3*time.Second {
			e.doorState = DS_Closeing
			e.doorStartTimer = time.Time{}
		}

	case DS_Closeing:
		if e.obstruction {
			e.doorState = DS_Obstruction
			break
		}

		elevio.SetDoorOpenLamp(false)
		e.doorState = DS_Closed

	case DS_Obstruction:
		elevio.SetDoorOpenLamp(true)

		if !e.obstruction {
			e.doorState = DS_Open
			e.doorStartTimer = time.Now()
		}

	case DS_Error:
	}
}

func (e *Elevator) RunDoorStateMachine() {
	fmt.Println("DOOR STATE MACHINE STARTED")
	ticker := time.NewTicker(50 * time.Millisecond)
	defer ticker.Stop()

	for range ticker.C {
		e.updateDoorState()
	}
}

// region printing
func (s DoorState) String() string {
	switch s {
	// case Idle:
	// 		return "idle"
	case DS_Closed:
		return "closed"
	case DS_Closeing:
		return "closeing"
	case DS_Open:
		return "open"
	case DS_Opening:
		return "opening"
	case DS_Obstruction:
		return "obstruction"
	case DS_Error:
		return "error"
	default:
		return "unknown"
	}
}

// endregion
