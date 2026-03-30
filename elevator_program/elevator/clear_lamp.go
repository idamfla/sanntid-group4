package elevator

import "elevator_program/elevio"

func (e *Elevator) clearCabLamp(floor int) {
	elevio.SetButtonLamp(elevio.BT_Cab, floor, false)
}

func (e *Elevator) clearHallLamp(floor int, button elevio.ButtonType) {
	elevio.SetButtonLamp(button, floor, false)
}

func (e *Elevator) clearAllLamps(buttons ...elevio.ButtonType) {
	for f := 0; f < e.NumFloors; f++ {
		for _, b := range buttons {
			elevio.SetButtonLamp(b, f, false)
		}
	}
}
