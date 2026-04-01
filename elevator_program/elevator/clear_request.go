package elevator

import (
	"elevator_program/elevio"
	"elevator_program/types"
)

func (e *Elevator) clearCabRequest(floor int) {
	e.System.Mutex.Lock()
	defer e.System.Mutex.Unlock()
	elevatorCopy := e.System.Elevators[e.Ip]
	elevatorCopy.CabRequests[floor] = types.NotActive
	e.System.Elevators[e.Ip] = elevatorCopy
}

func (e *Elevator) clearHallRequest(floor int, button elevio.ButtonType) {
	e.System.Mutex.Lock()
	defer e.System.Mutex.Unlock()
	e.System.HallRequests[floor][button] = types.NotActive
}

func (e *Elevator) clearCurrentFloor(floor int, button elevio.ButtonType) {
	if button == elevio.BT_Cab {
		e.clearCabRequest(floor)
		e.clearCabLamp(floor)
	} else {
		e.clearHallRequest(floor, button)
		e.clearHallLamp(floor, button)
	}
}
