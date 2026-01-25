package elevator

import (
	"elevator_program/elevio"
)

func (e *Elevator) clearRequestsForFloor(f int) {
	topFloor := len(e.floorRequests) - 1

	// always clear cab-request
	e.floorRequests[f][elevio.BT_Cab] = false

	switch f {
	case 0: 
		e.floorRequests[f][elevio.BT_HallUp] = false
	case topFloor: 
		e.floorRequests[f][elevio.BT_HallDown] = false
	default: 
		switch e.lastDirection {
		case elevio.MD_Up:
			e.floorRequests[f][elevio.BT_HallUp] = false
		case elevio.MD_Down:
			e.floorRequests[f][elevio.BT_HallDown] = false
		}
	}
}

func (e *Elevator) clearLampsForFloor(f int) {
	topFloor := len(e.floorRequests) - 1

	// always turn off cab-button
	elevio.SetButtonLamp(elevio.BT_Cab, e.currentFloor, false)

	switch f {
	case 0:
		elevio.SetButtonLamp(elevio.BT_HallUp, f, false)
	case topFloor:
		elevio.SetButtonLamp(elevio.BT_HallDown, f, false)
	default:
		switch e.lastDirection {
		case elevio.MD_Up:
			elevio.SetButtonLamp(elevio.BT_HallUp, e.currentFloor, false)
		case elevio.MD_Down:
			elevio.SetButtonLamp(elevio.BT_HallDown, e.currentFloor, false)
		}
	}
}

func (e *Elevator) clearCurrentFloor() {
	e.clearRequestsForFloor(e.currentFloor) // TODO don't clear floor before "master" tells the elevator to do so
	e.clearLampsForFloor(e.currentFloor)
}

func (e *Elevator) UpdateLastDirection(nextTarget elevio.ButtonEvent, dir elevio.MotorDirection) {
	if nextTarget.Floor == -1 {
		return
	}

	if dir != elevio.MD_Stop {
		e.lastDirection = dir
	}

	if nextTarget.Floor == e.currentFloor {
		switch nextTarget.Button {
		case elevio.BT_HallUp:
			e.lastDirection = elevio.MD_Up
		case elevio.BT_HallDown:
			e.lastDirection = elevio.MD_Down
		}
	}
}

func (e *Elevator) UpdateTargetFloor() elevio.ButtonEvent { // TODO should this return or just update?
	nextTarget := e.GetNextTargetFloor()
	if nextTarget.Floor != -1 {
		e.targetFloor = nextTarget.Floor
	}

	return nextTarget
}
