package elevator

import (
	"elevator_program/elevio"
	"elevator_program/utilities"
)

// TODO Chat saying this MSG_T_ naming convention is very c -style and noisy in Go
type MessageType int

const (
	MSG_T_StatusReport MessageType = iota

	MSG_T_TaskCreate   // a new task is created/published
	MSG_T_TaskAssign   // a task is assigned to you
	MSG_T_TaskDelegate // a task is assigned to another person
	MSG_T_TaskUpdate   // task changed, Don't think we need it
	MSG_T_TaskComplete // task was completed
	MSG_T_TaskRequest  // someone requests a new task
	MSG_T_LostComs     // A routine to check if you have lost communication
	MSG_T_NewToChannel // Send the latest information
)

type DecodedMessage struct {
	msgType        MessageType
	senderId       int
	task           elevio.ButtonEvent // Elevators current target (floor, btnType) or change current target
	btnStatus      ButtonStatus       // Type what we want the button to be: nonActive, pending, active
	elevatorStatus ElevatorsStatus

	// TODO temp need a com number
	comNumber int
	// TODO temp Need to know who the mission is to
	idToElevatorMission int

	// TODO Maybe we need target id as well

	// Used for a full sync
	fullstate *System
	// msgTimer       time.Time
	// TODO how to be able to send their chan Message as well
}

func decodeMessage(msg utilities.Message) DecodedMessage {
	// TODO probably don't need this
	// newMessage := DecodedMessage{
	// 	msgType:  MessageType(msg.MsgType),
	// 	senderId: msg.SenderId,
	// 	task: elevio.ButtonEvent{
	// 		Floor:  msg.Task.Floor,
	// 		Button: msg.Task.Button,
	// 	},
	// 	btnStatus: ButtonStatus(msg.BtnStatus),
	// 	elevatorStatus: ElevatorsStatus{
	// 		Id:           msg.ElevatorStatus.Id,
	// 		Ip:           msg.ElevatorStatus.Ip,
	// 		CurrentFloor: msg.ElevatorStatus.CurrentFloor,
	// 		CabRequests:  msg.ElevatorStatus.CabRequests,
	// 		Target: elevio.ButtonEvent{
	// 			Floor:  msg.ElevatorStatus.Target.Floor,
	// 			Button: msg.ElevatorStatus.Target.Button,
	// 		},
	// 		State: ElevatorState(msg.ElevatorStatus.State),
	// 	},
	// 	comNumber:           msg.ComNumber,
	// 	idToElevatorMission: msg.IdToElevatorMission,
	// 	fullstate: &System{
	// 		Elevators:    msg.Fullstate.Elevators,
	// 		hallRequests: msg.Fullstate.HallRequests,
	// 	},
	// }
	// return newMessage

}
