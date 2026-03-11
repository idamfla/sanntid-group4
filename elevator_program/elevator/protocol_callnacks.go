package elevator

import (
	"elevator_program/elevio"
	"elevator_program/message"
	"elevator_program/types"
)

// TODO This function is wierd, either we need to have it as e or something else if it is msg sending
func (e *Elevator) HandleLostConnection(senderId int) {
	// We need a way to know who initiatet the msg:
	// if we initiated the msg{
	if senderId == e.IpRegistery["Master"] {
		e.connectedToMaster = true
		// Make the msg resolved
	} else {
		// Need to count if any elevator har connection to master
		// Need to count how many does not have connection
		// if you are the problem, restart or complete cab then restart
		// If you are not the problem, select a new master
	}
}

func (e Elevator) UpdateBtnLamp(btnStatus types.ButtonStatus, floor int, button elevio.ButtonType) {
	if types.ButtonStatus(btnStatus) == types.NotActive {
		e.clearHallLamp(floor, button)
	} else {
		elevio.SetButtonLamp(button, floor, true)
	}
}

func (e *Elevator) SetConnectionState(msg message.Message) {
	e.id = msg.Id
	e.isMaster = false
	e.connectedToMaster = true
	e.elevatorState = types.ES_Idle // TODO Do I need this one here?
	// e.ipToId Need to know the ip/id to the others
}
