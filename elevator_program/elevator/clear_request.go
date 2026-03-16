package elevator

import (
	"elevator_program/elevio"
	"elevator_program/types"
)

func (e *Elevator) clearCabRequest(floor int) {
	e.System.Elevators[e.Id].CabRequests[floor] = types.NotActive // Changed to be compatible with system struct
}

func (e *Elevator) clearHallRequest(floor int, button elevio.ButtonType) {
	e.System.HallRequests[floor][button] = types.NotActive // Changed to be compatible with system struct
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
