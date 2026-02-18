package elevator

import (
	"elevator_program/elevio"
)

func (e *Elevator) clearCabRequest(floor int) { e.cabRequests[floor] = false }

func (e *Elevator) clearHallRequest(floor int, button elevio.ButtonType) {
	e.hallRequests[floor][button] = false
}

func (e Elevator) clearCabLamp(floor int) {
	elevio.SetButtonLamp(elevio.BT_Cab, floor, false)
}

func (e Elevator) clearHallLamp(floor int, button elevio.ButtonType) {
	elevio.SetButtonLamp(button, floor, false)
}

func (e Elevator) clearAllLamps(buttons ...elevio.ButtonType) {
	numFloors := len(e.hallRequests)
	for f := 0; f < numFloors; f++ {
		for _, b := range buttons {
			elevio.SetButtonLamp(b, f, false)
		}
	}
}

// Clear current floor from hallRequests, and turn the lamps off
func (e *Elevator) clearCurrentFloor(floor int, button elevio.ButtonType) {
	e.clearCabRequest(floor) // TODO don't clear floor before "master" tells the elevator to do so
	e.clearHallRequest(floor, button)

	e.clearCabLamp(floor)
	e.clearHallLamp(floor, button)
}
