package elevator

import (
	"elevator_program/elevio"
	"elevator_program/types"
	"fmt"
	"time"
)

// ------------------------
// Motion helper functions
// ------------------------
func (e Elevator) atTargetFloor() bool {
	return e.currentFloor == e.nextTarget.Floor && !e.inBetweenFloors && e.currentFloor != -1
}

func (e Elevator) isTargetValid() bool {
	return e.nextTarget.Floor >= 0 && e.nextTarget.Floor < len(e.hallRequests)
}

func (e Elevator) getMotion(target int) elevio.MotorDirection {
	if e.atTargetFloor() || e.emergencyStop || !e.isTargetValid() {
		return elevio.MD_Stop
	} else if e.currentFloor < target {
		return elevio.MD_Up
	} else {
		return elevio.MD_Down
	}
}

func (e *Elevator) updateDirection(target elevio.ButtonEvent, dir elevio.MotorDirection) {
	if dir == elevio.MD_Stop && target.Floor == e.currentFloor {
		switch target.Button {
		case elevio.BT_HallUp:
			dir = elevio.MD_Up
		case elevio.BT_HallDown:
			dir = elevio.MD_Down
		}
	}

	if dir != elevio.MD_Stop {
		e.direction = dir
	}
}

func (e Elevator) computeNextTargetAndDirection() (elevio.ButtonEvent, elevio.MotorDirection) {
	nextTarget := getNextTargetFloor(e)
	if nextTarget.Floor == -1 {
		return elevio.ButtonEvent{Floor: -1}, elevio.MD_Stop
	}

	dir := e.getMotion(nextTarget.Floor)
	return nextTarget, dir
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

	if e.elevatorState != types.ES_Uninitialized && e.doorState != DS_Closed {
		elevio.SetMotorDirection(elevio.MD_Stop)
		return
	}

	switch e.elevatorState {
	case types.ES_Uninitialized:
		dir = e.uninitializedAction()

		if dir == elevio.MD_Stop {
			e.clearCurrentFloor(e.currentFloor, elevio.BT_Cab)
			e.elevatorState = types.ES_Idle
			e.doorState = DS_Opening
			fmt.Println(e)
		}

	case types.ES_Idle:
		nextTarget, dir = e.computeNextTargetAndDirection()
		if nextTarget.Floor != -1 {
			e.nextTarget = nextTarget
			e.updateDirection(nextTarget, dir)
		}

		if e.atTargetFloor() { // TODO is it here bc if someone spams the button on the floor you're at?
			// e.doorState = open
			e.clearCurrentFloor(e.currentFloor, e.nextTarget.Button)
		}

		dir = e.getMotion(e.nextTarget.Floor)
		if dir != elevio.MD_Stop {
			e.elevatorState = types.ES_Moving
		}

	case types.ES_Moving:
		dir = e.getMotion(e.nextTarget.Floor)

		if dir == elevio.MD_Stop {
			e.doorState = DS_Opening
			e.elevatorState = types.ES_Idle
		} else {
			nextTarget, dir = e.computeNextTargetAndDirection()
			if nextTarget.Floor != -1 {
				// TODO I don't know if this is the best way to write it but now can use running
				if e.nextTarget.Button == elevio.BT_Cab {
					e.System.Elevators[e.id].CabRequests[e.nextTarget.Floor] = types.Pending // TODO Need to message that the buttons have changed
				} else {
					e.System.HallRequests[e.nextTarget.Floor][e.nextTarget.Button] = types.Pending // TODO Need to message that the buttons have changed
				}

				if nextTarget.Button == elevio.BT_Cab {
					e.System.Elevators[e.id].CabRequests[nextTarget.Floor] = types.Running
				} else {
					e.System.HallRequests[nextTarget.Floor][nextTarget.Button] = types.Running
				}
				e.nextTarget = nextTarget
				// e.hallRequests[nextTarget.Floor][nextTarget.Button] = Running
				e.updateDirection(nextTarget, dir)
			}
		}
	case types.ES_EmergencyStop:
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
