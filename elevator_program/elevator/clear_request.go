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
	e.clearCabLamp(floor)

	if floor == 0 {
		e.clearHallRequest(floor, elevio.BT_HallUp)
		e.clearHallLamp(floor, elevio.BT_HallUp)
		return

	} else if floor == len(e.hallRequests)-1 {
		e.clearHallRequest(floor, elevio.BT_HallDown)
		e.clearHallLamp(floor, elevio.BT_HallDown)
		return
	}

	if button == elevio.BT_Cab {
		return
	}
	e.clearHallRequest(floor, button)
	e.clearHallLamp(floor, button)
}
