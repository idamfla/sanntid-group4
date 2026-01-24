package elevator

import (
	"elevator_program/elevio"
)

func (e *Elevator) ClearFloor(f int) {
	e.floorRequests[f][elevio.BT_Cab] = false

	if f == 0 {
		e.floorRequests[f][elevio.BT_HallUp] = false
	} else if f == len(e.floorRequests)-1 {
		e.floorRequests[f][elevio.BT_HallDown] = false
	} else {
		switch e.lastDirection {
		case elevio.MD_Up:
			e.floorRequests[f][elevio.BT_HallUp] = false
		case elevio.MD_Down:
			e.floorRequests[f][elevio.BT_HallDown] = false
		}
	}
}

func (e *Elevator) ClearButtonLamp(f int) {
	elevio.SetButtonLamp(elevio.BT_Cab, e.currentFloor, false)

	if f == 0 {
		elevio.SetButtonLamp(elevio.BT_HallUp, f, false)
	} else if f == len(e.floorRequests)-1 {
		elevio.SetButtonLamp(elevio.BT_HallDown, f, false)
	} else {
		switch e.lastDirection {
		case elevio.MD_Up:
			elevio.SetButtonLamp(elevio.BT_HallUp, e.currentFloor, false)
		case elevio.MD_Down:
			elevio.SetButtonLamp(elevio.BT_HallDown, e.currentFloor, false)
		}
	}
}

func (e *Elevator) UpdateLastDirection(nextTarget elevio.ButtonEvent, dir elevio.MotorDirection) {
	if nextTarget.Floor == -1 {
		return
	}

	if dir != elevio.MD_Stop {
		e.lastDirection = dir
	}

	if nextTarget.Floor != e.currentFloor {
		return
	}
	switch nextTarget.Button {
	case elevio.BT_HallUp:
		e.lastDirection = elevio.MD_Up
	case elevio.BT_HallDown:
		e.lastDirection = elevio.MD_Down
	}
}

func (e *Elevator) UpdateTargetFloor() elevio.ButtonEvent {
	nextTarget := e.GetNextTargetFloor()
	if nextTarget.Floor != -1 {
		e.targetFloor = nextTarget.Floor
	}
	return nextTarget
}
