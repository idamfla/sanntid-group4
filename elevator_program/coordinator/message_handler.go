package coordinator

import (
	"elevator_program/elevator"
	"elevator_program/elevio"
	"elevator_program/message"
	"elevator_program/types"
	"elevator_program/udp/session"
	"fmt"
)

// Read new message from server when it appears on the channel
func (c *Coordinator) MessageListener(e *elevator.Elevator) {
	fmt.Println("MESSAGE LISTENER STARTED", e.Id)
	for ePkt := range c.msgRecieveCh {
		eMsg := ePkt.EMsg
		fmt.Println("Before \n\n\n\n\n\n", e.Id, e.IsMaster, e, eMsg)
		c.MessageHandler(e, eMsg)
		fmt.Println("Elevator after msg: ", e.Id, e.IsMaster, e, eMsg)
		if ePkt.Done != nil {
			ePkt.Done <- struct{}{}
		}
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

func (c *Coordinator) handleAsSlave(e *elevator.Elevator, eMsg message.ElevatorMessage) {
	switch eMsg.EMsgType {
	case message.EMSG_T_StatusReport:
		e.System.SetStatusReport(eMsg.ID, eMsg.Elevators[eMsg.ID])

	case message.EMSG_T_TaskUpdate:
		if e.Id == eMsg.ID && eMsg.BtnStatus == types.Running {
			e.System.SetRequestAsTarget(eMsg.ID, eMsg.Task)
		} else {
			e.System.Mutex.Lock()
			e.System.SetRequestStatus(eMsg.ID, eMsg.BtnStatus, eMsg.Task)
			e.System.Mutex.Unlock()
		}
		e.UpdateBtnLamp(eMsg.BtnStatus, eMsg.Task.Floor, eMsg.Task.Button)

	case message.EMSG_T_LostComs:
		if !e.ConnectedToMaster() {
			e.HandleLostConnection(eMsg.ID)
		}

	case message.EMSG_T_ElevatorLost:
		if e.ConnectedToMaster() {
			eMsg.ID = "" // Send "" if connected, TODO kind of wierd to send the value on Id
		} else {
			eMsg.ID = e.Id
		}
		eMsg.EMsgType = message.EMSG_T_LostComs // TODO Lost coms and Elevator lost should be deleted
		e.SendToCoordinator <- eMsg

	case message.EMSG_T_NewToChannel:
		if e.ConnectedToMaster() {
			e.IpRegistery[eMsg.Addr] = eMsg.ID // TODO now we can update IpRegistery for the others as well, is it smart?
			e.System.SetStatusReport(eMsg.ID, eMsg.Elevators[eMsg.ID])
		} else if e.Id == eMsg.ID {
			e.SetConnectionState(eMsg)
			e.System.InitializeFromSystemState(eMsg)
		} else {
			// // When two not connected elevators reach eachother
			// // The one with smallest ip gets to be master
			// senderIdInt, _ := strconv.Atoi(eMsg.ID)
			// ownIdInt, _ := strconv.Atoi(e.Id)
			// if ownIdInt < senderIdInt { // TODO It may be an error here if master sends back and another new elevator listens to it
			// 	e.TurnToMaster()
			// 	c.portRegistery["master"] = c.portSelf

			// 	msg, id := e.System.RegisterAndSyncElevator(eMsg, e.IpRegistery)
			// 	fmt.Println("Do I get here?", msg)
			// 	e.IpRegistery[eMsg.Addr] = id

			// 	e.SendToCoordinator <- msg
			// 	return
			// } else {
			// 	// You are not the master, continiue
			// 	return
			// }
		}
	}
}

func (c *Coordinator) handleAsMaster(e *elevator.Elevator, eMsg message.ElevatorMessage) {
	switch eMsg.EMsgType {
	case message.EMSG_T_StatusReport:
		e.System.SetStatusReport(eMsg.ID, eMsg.Elevators[eMsg.ID])
		e.SendToCoordinator <- eMsg

	case message.EMSG_T_ButtonPress:
		if eMsg.BtnStatus != types.NotActive {
			e.System.Mutex.RLock()
			_, elevs := e.System.Snapshot()
			e.System.Mutex.RUnlock()

			if eMsg.Task.Button == elevio.BT_Cab {
				if e.IsNewTargetBetterCab(eMsg.ID, eMsg.Task, elevs[eMsg.ID]) {
					eMsg.BtnStatus = types.Running
				} else {
					eMsg.BtnStatus = types.Pending
				}
			} else {
				taskElevatorId, _, _ := e.ClosestToTarget(elevs, eMsg.Task) // TODO could be wrong here if master don't update system
				if taskElevatorId != "" {
					// Someone has a better task to do, we need to broadcast task_Update
					eMsg.ID = taskElevatorId
					eMsg.BtnStatus = types.Running
				} else { // If it is not the case we just need to broadcast the change
					eMsg.ID = ""
				}
				eMsg.EMsgType = message.EMSG_T_ButtonPress //MSG_T_TaskUpdate
			}
		}
		e.SendToCoordinator <- eMsg

		eMsg.EMsgType = message.EMSG_T_TaskUpdate // TODO should remove this

		packet := session.ElevatorPacket{
			EMsg: eMsg,
		}

		c.msgRecieveCh <- packet

	case message.EMSG_T_TaskUpdate:
		e.System.SetRequestStatus(eMsg.ID, eMsg.BtnStatus, eMsg.Task)
		if !(eMsg.ID == e.Id && eMsg.Task.Button == elevio.BT_Cab) {
			e.UpdateBtnLamp(eMsg.BtnStatus, eMsg.Task.Floor, eMsg.Task.Button)
		}

		taskKey := TaskKey{
			Owner:  eMsg.ID,
			TaskID: eMsg.Task,
		}
		switch eMsg.BtnStatus {
		case types.Running:
			c.TaskMonitor.StartTask(taskKey, e)
		case types.NotActive:
			c.TaskMonitor.FinishTask(taskKey)
		}

	case message.EMSG_T_TaskRequest:
		e.System.Mutex.RLock()
		hallRequests, elevs := e.System.Snapshot() // TODO changed to snapshot, did any new errors apear
		e.System.Mutex.RUnlock()
		task := e.GetNextTargetFloor(elevs[eMsg.ID], hallRequests)
		if task.Floor != -1 {
			// Broadcast new assignment if we found a new task
			eMsg.Task = task
			eMsg.BtnStatus = types.Running
			e.SendToCoordinator <- eMsg
		}

	case message.EMSG_T_NewToChannel:
		eMsg, id := e.System.RegisterAndSyncElevator(eMsg, e.IpRegistery)
		e.IpRegistery[eMsg.Addr] = id
		e.SendToCoordinator <- eMsg
	}
}
