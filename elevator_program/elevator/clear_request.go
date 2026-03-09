package elevator

import (
	"elevator_program/elevio"
)

func (e *Elevator) clearCabRequest(floor int) {
	e.system.Elevators[e.id].CabRequests[floor] = NotActive // Changed to be compatible with system struct
}

func (e *Elevator) clearHallRequest(floor int, button elevio.ButtonType) {
	// println("the length here is: ", len(e.hallRequests[3]))
	// println("floor: ", floor)
	// println("Button: ", button)
	e.system.hallRequests[floor][button] = NotActive // Changed to be compatible with system struct
}

// Clear current floor from hallRequests, and turn the lamps off
func (e *Elevator) clearCurrentFloor(floor int, button elevio.ButtonType) {
	e.clearCabRequest(floor) // TODO don't clear floor before "master" tells the elevator to do so
	e.clearCabLamp(floor)

	if button == elevio.BT_Cab {
		return
	}

	e.clearHallRequest(floor, button)
	e.clearHallLamp(floor, button)
}
