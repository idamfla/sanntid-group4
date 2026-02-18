package elevator

import (
	"elevator_program/elevio"
)

func (e *Elevator) clearCabRequest(floor int) { e.cabRequests[floor] = false }

func (e *Elevator) clearHallRequest(floor int, button elevio.ButtonType) {
	e.hallRequests[floor][button] = false
}

// Clear current floor from hallRequests, and turn the lamps off
func (e *Elevator) clearCurrentFloor(floor int, button elevio.ButtonType) {
	e.clearCabRequest(floor) // TODO don't clear floor before "master" tells the elevator to do so
	e.clearHallRequest(floor, button)

	e.clearCabLamp(floor)
	e.clearHallLamp(floor, button)
}
