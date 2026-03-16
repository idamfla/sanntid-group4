package protocol

import (
	"elevator_program/elevator"
	"elevator_program/message"
	"elevator_program/types"
	"fmt"
	"strconv"
)

// TODO Ida thinks it could be a better way to structure it

/*
TODO Needs this somewhere
ctx, cancel := context.Withcancel(context.Background())
msgCh := make(chan string)
*/

// Chat don't like these function names. Don't want any underscores
func (p *Protocol) slaveMessageHandler(e *elevator.Elevator, msg message.Message) {
	switch msg.MsgType {
	case types.MSG_T_StatusReport:
		e.System.SetStatusReport(msg.Id, msg.Elevators[msg.Id])
		fmt.Println("Status rapport: ", e.System)

	case types.MSG_T_TaskUpdate:
		if e.Id == msg.Id && msg.BtnStatus == types.Running { // Assign new task
			e.SetRequestAsTarget(msg.Task)
		} else { // Just update system
			e.System.SetRequestStatus(msg.Id, msg.BtnStatus, msg.Task) // TODO Should I use my own ID or msg Id??
		}
		e.UpdateBtnLamp(msg.BtnStatus, msg.Task.Floor, msg.Task.Button)

	case types.MSG_T_LostComs:
		e.HandleLostConnection(msg.Id)

	case types.MSG_T_ElevatorLost:
		isConnectedToMaster := e.ConnectedToMaster()
		if isConnectedToMaster {
			msg.Id = "" // Send "" if connected, TODO kind of wierd to send the value on Id
		} else {
			msg.Id = e.Id
		}
		msg.MsgType = types.MSG_T_LostComs
		fmt.Println("Message handler 1")
		e.SendToProtocol <- msg

	case types.MSG_T_NewToChannel:
		fmt.Println("Messaging: ", msg)
		if e.ConnectedToMaster() {
			e.IpRegistery[msg.Ip] = msg.Id // TODO now we can update IpRegistery for the others as well, is it smart?
			e.System.SetStatusReport(msg.Id, msg.Elevators[msg.Id])
		} else if e.Id == msg.Id {
			e.SetConnectionState(msg)
			e.System.InitializeFromSystemState(msg)
		} else {
			// When two not connected elevators reach eachother
			// The one with smallest ip gets to be master
			senderIdInt, _ := strconv.Atoi(msg.Id)
			ownIdInt, _ := strconv.Atoi(e.Id)
			if ownIdInt < senderIdInt {
				e.TurnToMaster()

				msg, id := e.System.RegisterAndSyncElevator(msg, e.IpRegistery)
				fmt.Println("Now i am going to send back that this one is connected to network: ", msg)
				e.IpRegistery[msg.Ip] = id
				fmt.Println("Message handler 2")
				e.SendToProtocol <- msg
			} else {
				// You are not the master, continiue
				return // TODO is it possible the last number is the same?
			}
		}
		p.activePeers = msg.ActivePeers
		p.Server.UpdateActivePeers(p.activePeers)
	}
}

// TODO chat don't like this name either
func (p *Protocol) masterMessageHandler(e *elevator.Elevator, msg message.Message) {
	switch msg.MsgType {
	case types.MSG_T_StatusReport:
		e.System.SetStatusReport(msg.Id, msg.Elevators[msg.Id])
		// TODO Send broadcast of status report
		fmt.Println("Message handler 3")
		e.SendToProtocol <- msg

	case types.MSG_T_ButtonPress:
		// TODO Could have a test to prevent duplicated requests, check if s.task == msg.BtnStatus

		taskElevatorId, _, _ := e.ClosestToTarget(e.System.Elevators, msg.Task)
		fmt.Println("aM i at buttonpress??")
		if taskElevatorId != "" {
			msg.MsgType = types.MSG_T_TaskUpdate
			msg.Id = taskElevatorId
			msg.BtnStatus = types.Running
			// Someone has a better task to do, we need to broadcast task_Update
		} else { // If it is not the case we just need to broadcast the change
			msg.MsgType = types.MSG_T_TaskUpdate
			msg.Id = ""
			msg.BtnStatus = types.Pending
		}
		fmt.Println("Message handler 4")
		e.SendToProtocol <- msg

	case types.MSG_T_TaskUpdate:
		e.System.SetRequestStatus(msg.Id, msg.BtnStatus, msg.Task)      // TODO These shouldn't be uppdated before everyone is ready
		e.UpdateBtnLamp(msg.BtnStatus, msg.Task.Floor, msg.Task.Button) // TODO SetRequestStatus can be there but not the light

	case types.MSG_T_TaskRequest:
		// Scan for the next request and send it back
		cabRequestsTemp := msg.Elevators[msg.Id].CabRequests
		currentFloor := msg.Elevators[msg.Id].CurrentFloor
		direction := msg.Elevators[msg.Id].Direction
		task := e.ComputeNewTarget(currentFloor, cabRequestsTemp, direction)
		fmt.Println("Need to send new task: ", task)
		if task.Floor != -1 {
			// Broadcast new assignment if we found a new task
			msg.Task = task
			msg.BtnStatus = types.Running
			fmt.Println("Message handler 5")
			e.SendToProtocol <- msg
		}

	case types.MSG_T_NewToChannel:
		p.activePeers++
		p.Server.UpdateActivePeers(p.activePeers)

		msg, id := e.System.RegisterAndSyncElevator(msg, e.IpRegistery)
		e.IpRegistery[msg.Ip] = id
		fmt.Println("Message handler 6")
		e.SendToProtocol <- msg
	}
}

func (p *Protocol) MessageHandler(e *elevator.Elevator, msg message.Message) {
	if e.IsMaster {
		p.masterMessageHandler(e, msg)
	} else {
		fmt.Println("5000 IDown, idmsg: ", e.Id, msg.Id)
		p.slaveMessageHandler(e, msg)
	}
}

func (p *Protocol) MessageListener(e *elevator.Elevator) {
	fmt.Println("MESSAGE LISTENER STARTED")
	for pktCtx := range p.msgRecieveCh {
		fmt.Println("Wallah broren min")
		msg := pktCtx.Packet.Payload
		fmt.Println("IDown, idmsg: ", e.Id, msg.Id)
		p.MessageHandler(e, msg)
		fmt.Println("Elevator after msg: ", e.Id, e.IsMaster, e.System)
		if pktCtx.Done != nil {
			pktCtx.Done <- struct{}{} // TODO this may not work
		}
	}
}

/*
------------------------------------------------------------------------------
Applying protocol functions which is ment to split between the different roles
------------------------------------------------------------------------------
*/

// func (p Protocol) applyStatusReport(e *elevator.Elevator, msg message.Message) {
// 	e.System.SetStatusReport(msg.Id, msg.Elevators[msg.Id])
// }

// func (p Protocol) applyTaskUpdate_slave(e *elevator.Elevator, msg message.Message) {
// 	e.System.SetRequestStatus(msg.Id, msg.BtnStatus, *msg.Task) // TODO Should I use my own ID or msg Id??
// 	e.UpdateBtnLamp(msg.BtnStatus, msg.Task.Floor, msg.Task.Button)
// }

// func (p Protocol) applyRemoteCabUpdate_slave(e *elevator.Elevator, msg message.Message) {
// 	e.System.UpdateRemoteCabBtn(msg.Id, msg.BtnStatus, msg.Task.Floor)
// }

// func (p Protocol) applyLostComsProtocol_slave(e *elevator.Elevator, msg message.Message) {
// 	e.HandleLostConnection(msg.Id)
// }

// func (p Protocol) applySystemSync_slave(e *elevator.Elevator, msg message.Message) {
// 	e.System.InitializeFromSystemState(msg)
// 	e.SetConnectionState(msg)
// }

// func (p *Protocol) addNewRequestToSystem_master(e *elevator.Elevator, msg message.Message) {
// 	// TODO Lets hope that we only get commit messages, or else we need to count ack
// 	taskElevatorId, _, _ := e.ClosestToTarget(e.System.Elevators, *msg.Task)
// 	if taskElevatorId != -1 {

// 	}
// 	e.System.SetRequestStatus(msg.Id, msg.BtnStatus, *msg.Task)
// 	e.UpdateBtnLamp(msg.BtnStatus, msg.Task.Floor, msg.Task.Button)
// }

// func (p *Protocol) applyRegisterAndSyncElevatorToServer(e *elevator.Elevator, msg message.Message) {
// 	e.System.RegisterAndSyncElevator(msg, e.IpRegistery)
// }
