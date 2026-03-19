package elevator

import (
	"elevator_program/elevio"
	"elevator_program/message"
	"elevator_program/types"
)

// TODO This function is wierd, either we need to have it as e or something else if it is msg sending
func (e *Elevator) HandleLostConnection(senderId string) {
	if senderId == "" { //|| time.Since(e.lostComsTimer) > 4*time.Second
		e.IsOnline = false
		// Need to schedule a restart
		// TODO It is fault tolerance that should take the time maybe
	} else {
		e.ackCounterLostComs++
		e.System.Mutex.RLock()
		if e.ackCounterLostComs >= len(e.System.Elevators)-1 { // TODO something not right here, maybe master and a slave has died, then you are waiting for the dead slave to respond as well
			// Need to reset the timer
			// Need to start election
		}
		e.System.Mutex.RUnlock()
	}
}

func (e *Elevator) ConnectedToMaster() bool {
	return e.connectedToMaster
}

func (e *Elevator) UpdateBtnLamp(btnStatus types.ButtonStatus, floor int, button elevio.ButtonType) {
	if btnStatus == types.NotActive {
		if button == elevio.BT_Cab {
			e.clearCabLamp(floor)
		} else {
			e.clearHallLamp(floor, button) // Chat don't like these function names. Don't want any underscores
		}

	} else {
		elevio.SetButtonLamp(button, floor, true)
	}
}

func (e *Elevator) SetConnectionState(msg message.ElevatorMessage) {
	e.Id = msg.Id
	e.IsMaster = false
	e.connectedToMaster = true
	e.IsOnline = true
	e.System.Mutex.RLock()
	for id, elevator := range e.System.Elevators {
		e.IpRegistery[elevator.Ip] = id
	}
	e.System.Mutex.RUnlock()
}

// TODO Probably don't need, just for testing
// func (e *Elevator) ClearElevator(numFloors int) {
// 	e.System.HallRequests = make([][2]types.ButtonStatus, numFloors)
// 	e.System.Elevators = make(map[string]types.ElevatorsStatus)
// 	// e.nextTarget = elevio.ButtonEvent{
// 	// 	Floor:  -1,
// 	// 	Button: elevio.BT_HallUp,
// 	// }
// 	elevio.SetMotorDirection(0)
// 	// TODO if i want to test this one, have to change to systemstate
// 	// e.elevatorState = types.ES_EmergencyStop
// }

func (e *Elevator) ClearTarget() {
	clearedTarget := elevio.ButtonEvent{
		Floor:  -1,
		Button: elevio.BT_HallUp,
	}
	e.System.Mutex.Lock()
	elevatorCopy := e.System.Elevators[e.Id]
	elevatorCopy.Target = clearedTarget
	e.System.Elevators[e.Id] = elevatorCopy
	e.System.Mutex.Unlock()
}

func (e *Elevator) TurnToMaster() {
	e.IsOnline = true
	e.IsMaster = true
	e.connectedToMaster = true
}
