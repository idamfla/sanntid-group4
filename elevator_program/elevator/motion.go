package elevator

import(
	"elevator_program/elevio"
)

func (e Elevator) getMotion() elevio.MotorDirection {
	if e.targetFloor == -1 || e.currentFloor == e.targetFloor || e.emergencyStop {
		return elevio.MD_Stop
	} else if e.currentFloor < e.targetFloor {
		return elevio.MD_Up
	} else {
		return elevio.MD_Down
	}
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

func (e *Elevator) updateMotor() {
	if e.emergencyStop { 
		elevio.SetMotorDirection(elevio.MD_Stop)
		return
	}

	var dir elevio.MotorDirection = elevio.MD_Stop

	switch e.state {
	case ES_Uninitialized:
		dir = e.uninitializedAction()

		if dir == elevio.MD_Stop {
			e.clearCurrentFloor()
			e.state = ES_Idle
		}

	case ES_Idle:
		nextTarget := e.UpdateTargetFloor() // sets targetFloor, TODO does this need to be a function or can i do it directly
		dir = e.getMotion()

		if dir != elevio.MD_Stop {
			e.state = ES_Moving
		}

		if e.currentFloor != -1 && e.currentFloor == e.targetFloor { // TODO is it here bc if someone spams the button on the floor you're at?
			e.clearCurrentFloor()
		}

		e.UpdateLastDirection(nextTarget, dir)

	case ES_Moving:
		if e.currentFloor == e.targetFloor {
			e.clearCurrentFloor()
			dir = elevio.MD_Stop
			e.state = ES_Idle // TODO state to doorOpen
		} else {
			nextTarget := e.UpdateTargetFloor()
			dir = e.getMotion()
			e.UpdateLastDirection(nextTarget, dir)
		}
	}

	elevio.SetMotorDirection(dir)
}