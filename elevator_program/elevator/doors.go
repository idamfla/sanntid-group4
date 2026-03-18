package elevator

import (
	"elevator_program/elevio"
	"fmt"
	"time"
)

type DoorState int

const (
	DS_Closed DoorState = iota
	DS_Closing
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

		e.System.Mutex.RLock()
		task := e.System.Elevators[e.Id].Target
		e.System.Mutex.RUnlock()
		if e.atTargetFloor(task.Floor) {
			e.clearCurrentFloor(e.currentFloor, task.Button)
		}

		e.doorState = DS_Open
		e.doorTimer = time.Time{}

	case DS_Open:
		if e.obstruction {
			e.doorState = DS_Obstruction
			e.doorTimer = time.Time{} // reset timer
			break
		}

		if e.doorTimer.IsZero() {
			e.doorTimer = time.Now()
		}

		if time.Since(e.doorTimer) >= 3*time.Second {
			e.doorState = DS_Closing
			e.doorTimer = time.Time{}
		}

	case DS_Closing:
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
			e.doorTimer = time.Now()
		}

	case DS_Error:
	}
}

func (e *Elevator) RunDoorStateMachine() {
    defer e.wg.Done()
	fmt.Println("DOOR STATE MACHINE STARTED")
	ticker := time.NewTicker(50 * time.Millisecond)
	defer ticker.Stop()

	for {
        select {
        case <-e.stop:
            return
        case <-ticker.C:
            e.updateDoorState()
        }
    }
}

// region printing
func (s DoorState) String() string {
	switch s {
	// case Idle:
	// 		return "idle"
	case DS_Closed:
		return "closed"
	case DS_Closing:
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
