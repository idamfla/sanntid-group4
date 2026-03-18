package elevator

import "elevator_program/elevio"

// TODO move to a hardware map

func (e *Elevator) clearCabLamp(floor int) { // TODO Why do we need to separate between cab and hall??
	elevio.SetButtonLamp(elevio.BT_Cab, floor, false)
}

func (e *Elevator) clearHallLamp(floor int, button elevio.ButtonType) {
	elevio.SetButtonLamp(button, floor, false)
}

func (e *Elevator) clearAllLamps(buttons ...elevio.ButtonType) {
	numFloors := len(e.hallRequests)
	for f := 0; f < numFloors; f++ {
		for _, b := range buttons {
			elevio.SetButtonLamp(b, f, false)
		}
	}
}
