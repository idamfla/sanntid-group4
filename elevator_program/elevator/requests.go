package elevator

import (
	"elevator_program/elevio"
)

// Helper function, do not call directly, use (e *Elevator) clearCurrentFloor
func (e *Elevator) clearRequestsForFloor(floor int) {
	topFloor := len(e.floorRequests) - 1

	// always clear cab-request
	e.cabRequests[floor] = false

	switch floor {
	case 0:
		e.floorRequests[floor][elevio.BT_HallUp] = false
	case topFloor:
		e.floorRequests[floor][elevio.BT_HallDown] = false
	default:
		switch e.lastDirection {
		case elevio.MD_Up:
			e.floorRequests[floor][elevio.BT_HallUp] = false
		case elevio.MD_Down:
			e.floorRequests[floor][elevio.BT_HallDown] = false
		}
	}
}

// Helper function, do not call directly, use (e *Elevator) clearCurrentFloor
func (e Elevator) clearLampsForFloor(floor int) {
	topFloor := len(e.floorRequests) - 1

	// always turn off cab-button
	elevio.SetButtonLamp(elevio.BT_Cab, e.currentFloor, false)

	switch floor {
	case 0:
		elevio.SetButtonLamp(elevio.BT_HallUp, floor, false)
	case topFloor:
		elevio.SetButtonLamp(elevio.BT_HallDown, floor, false)
	default:
		switch e.lastDirection {
		case elevio.MD_Up:
			elevio.SetButtonLamp(elevio.BT_HallUp, e.currentFloor, false)
		case elevio.MD_Down:
			elevio.SetButtonLamp(elevio.BT_HallDown, e.currentFloor, false)
		}
	}
}

func (e Elevator) clearAllLamps(buttons ...elevio.ButtonType) {
	numFloors := len(e.floorRequests)
	for f := 0; f < numFloors; f++ {
		for _, b := range buttons {
			elevio.SetButtonLamp(b, f, false)
		}
	}
}

// Clear current floor from floorRequests, and turn the lamps off
func (e *Elevator) clearCurrentFloor() {
	e.clearRequestsForFloor(e.currentFloor) // TODO don't clear floor before "master" tells the elevator to do so
	e.clearLampsForFloor(e.currentFloor)
}
