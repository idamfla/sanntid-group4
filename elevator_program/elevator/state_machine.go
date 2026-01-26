package elevator

import (
	"fmt"
	"time"
	"elevator_program/elevio"
)

type ElevatorState int

const (
	ES_Uninitialized ElevatorState = iota
	ES_Idle
	ES_Moving
	ES_DoorOpen
	// ES_Obstruction // TODO move to DoorState
	ES_EmergencyStop
)

// ------------------------
// Motion helper functions
// ------------------------
func (e Elevator) atTargetFloor() bool {
	if e.currentFloor == -1 || e.nextTarget.Floor == -1 { // Failsafe, currentFloor should never be -1
		return false
	}

	return e.currentFloor == e.nextTarget.Floor && !e.inBetweenFloors
}

func (e Elevator) getMotionForTargetFloor(target int) elevio.MotorDirection {
	if e.atTargetFloor() || e.emergencyStop {
		return elevio.MD_Stop
	} else if e.currentFloor < target {
		return elevio.MD_Up
	} else {
		return elevio.MD_Down
	}
}

func (e Elevator) computeLastMovement(target elevio.ButtonEvent) elevio.MotorDirection {
	dir := e.getMotionForTargetFloor(target.Floor)

	if dir == elevio.MD_Stop && target.Floor == e.currentFloor {
		switch target.Button {
		case elevio.BT_HallUp:
			dir = elevio.MD_Up
		case elevio.BT_HallDown:
			dir = elevio.MD_Down
		}
	}

	return dir
}

func (e Elevator) uninitializedAction() elevio.MotorDirection {
	if e.currentFloor == -1 {
		return elevio.MD_Down
	}

	if e.currentFloor < e.initFloor {
		return elevio.MD_Up
	}

	if e.currentFloor > e.initFloor {
		return elevio.MD_Down
	}

	return elevio.MD_Stop
}


func (e *Elevator) updateLastDirection() {
	lastDir := e.computeLastMovement(e.nextTarget)
	if lastDir != elevio.MD_Stop {
		e.lastDirection = lastDir
	}
}

func (e Elevator) computeNextTargetAndDirection() (elevio.ButtonEvent, elevio.MotorDirection) {
    nextTarget := e.getNextTargetFloor()
    if nextTarget.Floor == -1 {
        return elevio.ButtonEvent{Floor: -1}, elevio.MD_Stop
    }

    dir := e.getMotionForTargetFloor(nextTarget.Floor)
    return nextTarget, dir
}



// ------------------------
// State Machine
// ------------------------
func (e *Elevator) updateElevatorState() { // TODO rename, this change state and controls the motor
	if e.emergencyStop { 
		elevio.SetMotorDirection(elevio.MD_Stop)
		return
	}

	// TODO add doorstate switch, e.startTime = time.Now()

	var dir elevio.MotorDirection = elevio.MD_Stop
	var nextTarget elevio.ButtonEvent = elevio.ButtonEvent{Floor: -1, Button: elevio.BT_Cab}

	if e.state != ES_Uninitialized && e.doorState != DS_Closed {
		elevio.SetMotorDirection(elevio.MD_Stop)
		return
	}

	switch e.state {
	case ES_Uninitialized:
		dir = e.uninitializedAction()

		if dir == elevio.MD_Stop {
			e.clearCurrentFloor()
			e.state = ES_Idle
			e.doorState = DS_Opening
			fmt.Println(e)
		}

	case ES_Idle:
		nextTarget, dir = e.computeNextTargetAndDirection()
		if nextTarget.Floor != -1 {
			e.nextTarget = nextTarget
			e.updateLastDirection()
		}

		if e.atTargetFloor() { // TODO is it here bc if someone spams the button on the floor you're at?
			// e.doorState = open
			e.clearCurrentFloor()
		}
	
		dir = e.getMotionForTargetFloor(e.nextTarget.Floor)
		if dir != elevio.MD_Stop {
			e.state = ES_Moving
		}

	case ES_Moving:
		dir = e.getMotionForTargetFloor(e.nextTarget.Floor)

		if dir == elevio.MD_Stop {
			e.doorState = DS_Opening
			e.state = ES_Idle
		} else {
			nextTarget, _ = e.computeNextTargetAndDirection()
			if nextTarget.Floor != -1 {
				e.nextTarget = nextTarget
				e.updateLastDirection()
			}
		}
	case ES_EmergencyStop:
		return
	}

	elevio.SetMotorDirection(dir)
}

func (e *Elevator) RunElevatorStateMachine() {
	fmt.Println("ELEVATOR STATE MACHINE STARTED")
	ticker := time.NewTicker(50 * time.Millisecond)
	defer ticker.Stop()

	for range ticker.C {
        e.updateElevatorState()
    }
}

// region printing
func (s ElevatorState) String() string {
	switch s {
	// case Idle:
	// 		return "idle"
	case ES_Uninitialized:
		return "uninitialized"
	case ES_Idle:
		return "idle"
	case ES_Moving:
		return "moving"
	case ES_DoorOpen:
		return "door open"
	// case ES_Obstruction:
	// 	return "obstruction"
	case ES_EmergencyStop:
		return "emergency stop"
	default:
		return "unknown"
	}
}
// endregion