package coordinator

import (
	"elevator_program/elevator"
	"elevator_program/elevio"
	"elevator_program/message"
	"elevator_program/types"
	"fmt"
	"strconv"
)

// TODO Chat thinks that this name is not that good, should use follower instead, but then we need to know that everyone else is also using this
func (c *Coordinator) handleAsSlave(e *elevator.Elevator, msg message.ElevatorMessage) {
	switch msg.MsgType {
	case types.MSG_T_StatusReport:
		e.System.SetStatusReport(msg.Id, msg.Elevators[msg.Id])

	case types.MSG_T_TaskUpdate:
		if e.Id == msg.Id && msg.BtnStatus == types.Running {
			e.System.SetRequestAsTarget(msg.Id, msg.Task)
		} else {
			e.System.Mutex.Lock()
			e.System.SetRequestStatus(msg.Id, msg.BtnStatus, msg.Task)
			e.System.Mutex.Unlock()
		}
		e.UpdateBtnLamp(msg.BtnStatus, msg.Task.Floor, msg.Task.Button)

	case types.MSG_T_LostComs:
		if !e.ConnectedToMaster() {
			e.HandleLostConnection(msg.Id)
		}

	case types.MSG_T_ElevatorLost:
		if e.ConnectedToMaster() {
			msg.Id = "" // Send "" if connected, TODO kind of wierd to send the value on Id
		} else {
			msg.Id = e.Id
		}
		msg.MsgType = types.MSG_T_LostComs
		e.SendToCoordinator <- msg

	case types.MSG_T_NewToChannel:
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
			if ownIdInt < senderIdInt { // TODO It may be an error here if master sends back and another new elevator listens to it
				e.TurnToMaster()
				c.portRegistery["master"] = c.portSelf

				msg, id := e.System.RegisterAndSyncElevator(msg, e.IpRegistery)
				fmt.Println("Do I get here?", msg)
				e.IpRegistery[msg.Ip] = id

				e.SendToCoordinator <- msg
				return
			} else {
				// You are not the master, continiue
				return
			}
		}
	}
}

func (c *Coordinator) handleAsMaster(e *elevator.Elevator, msg message.ElevatorMessage) {
	switch msg.MsgType {
	case types.MSG_T_StatusReport:
		e.System.SetStatusReport(msg.Id, msg.Elevators[msg.Id])
		// TODO Send broadcast of status report
		e.SendToCoordinator <- msg

	case types.MSG_T_ButtonPress:
		// TODO Could have a test to prevent duplicated requests, check if s.task == msg.BtnStatus
		if msg.BtnStatus != types.NotActive {
			if msg.Task.Button == elevio.BT_Cab {
				if e.IsNewTargetBetterCab(msg.Id, msg.Task, msg.Elevators[msg.Id]) {
					msg.BtnStatus = types.Running
				} else {
					msg.BtnStatus = types.Pending
				}
			} else {
				fmt.Println("What does master see? ", e.Id)
				e.System.Mutex.RLock()
				elevatorsCopy := e.System.Elevators
				e.System.Mutex.RUnlock()
				taskElevatorId, _, _ := e.ClosestToTarget(elevatorsCopy, msg.Task) // TODO could be wrong here if master don't update system
				if taskElevatorId != "" {
					// Someone has a better task to do, we need to broadcast task_Update
					msg.Id = taskElevatorId
					msg.BtnStatus = types.Running
				} else { // If it is not the case we just need to broadcast the change
					msg.Id = ""
				}
				msg.MsgType = types.MSG_T_ButtonPress //MSG_T_TaskUpdate
			}
		}
		e.SendToCoordinator <- msg

	case types.MSG_T_TaskUpdate:
		e.System.SetRequestStatus(msg.Id, msg.BtnStatus, msg.Task)
		if !(msg.Id == e.Id && msg.Task.Button == elevio.BT_Cab) {
			e.UpdateBtnLamp(msg.BtnStatus, msg.Task.Floor, msg.Task.Button)
		}

		taskKey := TaskKey{
			Owner:  msg.Id,
			TaskID: msg.Task,
		}
		switch msg.BtnStatus {
		case types.Running:
			c.TaskMonitor.StartTask(taskKey, e)
		case types.NotActive:
			c.TaskMonitor.FinishTask(taskKey)
		default:
			// TODO should i remove it? should not do anything here
		}

	case types.MSG_T_TaskRequest:
		e.System.Mutex.RLock()
		hallRequests := e.System.HallRequests
		e.System.Mutex.RUnlock()
		task := e.GetNextTargetFloor(msg.Elevators[msg.Id], hallRequests)
		fmt.Println("After computing \n\n ", task, msg.Elevators, msg.HallRequests, msg.Id)
		if task.Floor != -1 {
			// Broadcast new assignment if we found a new task
			msg.Task = task
			msg.BtnStatus = types.Running
			e.SendToCoordinator <- msg
		}

	case types.MSG_T_NewToChannel:
		msg, id := e.System.RegisterAndSyncElevator(msg, e.IpRegistery)
		e.IpRegistery[msg.Ip] = id
		e.SendToCoordinator <- msg
	}
}

// Route the message to a handler
func (c *Coordinator) MessageHandler(e *elevator.Elevator, msg message.ElevatorMessage) {
	if e.IsMaster {
		c.handleAsMaster(e, msg)
	} else {
		c.handleAsSlave(e, msg)
	}
}

// Read new message from server when it appears on the channel
func (c *Coordinator) MessageListener(e *elevator.Elevator) {
	fmt.Println("MESSAGE LISTENER STARTED", e.Id)
	for pktCtx := range c.msgRecieveCh {
		msg := pktCtx.Packet.Payload
		fmt.Println("Before \n\n\n\n\n\n", e.Id, e.IsMaster, e, msg)
		c.MessageHandler(e, msg)
		fmt.Println("Elevator after msg: ", e.Id, e.IsMaster, e, msg)
		if pktCtx.Done != nil {
			pktCtx.Done <- struct{}{}
		}
	}
}
