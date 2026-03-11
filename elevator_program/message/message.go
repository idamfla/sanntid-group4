package message

import (
	"elevator_program/elevio"
	"elevator_program/types"
)

// DTOs, data transfer objects

// Hopefully a better struct
type Message struct {
	MsgType types.MessageType

	Id int
	Ip string

	// Elevator state reporting
	// Status *types.ElevatorsStatus // TODO Don't think we need this one, only need to use Elevator map

	// Task / button updates
	Task      *elevio.ButtonEvent // TODO do we want it as a pointer? Gives us the option to not send Task on every message
	BtnStatus types.ButtonStatus

	// System synchronization
	HallRequests [][2]types.ButtonStatus
	Elevators    map[int]types.ElevatorsStatus
}
