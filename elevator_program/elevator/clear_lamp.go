package elevator

import "elevator_program/elevio"

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
