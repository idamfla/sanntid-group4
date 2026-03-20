package coordinator

import (
	"elevator_program/elevator"
	"elevator_program/elevio"
	"elevator_program/message"
	"elevator_program/types"
	"elevator_program/udp"
	"elevator_program/udp/packet"
	"fmt"
)

func (c *Coordinator) MessageListener(e *elevator.Elevator) {
	fmt.Println("MESSAGE LISTENER STARTED", e.Id)
	for ePkt := range c.msgRecieveCh {
		eMsg := ePkt.EMsg
		c.MessageHandler(e, eMsg)
		if ePkt.Done != nil {
			ePkt.Done <- struct{}{}
		}
	}
}

func (c *Coordinator) MessageHandler(e *elevator.Elevator, msg message.ElevatorMessage) {
	if e.IsMaster {
		e.System.Mutex.Lock()
		e.IsMaster = true
		e.System.Mutex.RUnlock()
		c.handleAsMaster(e, msg)
	} else {
		c.handleAsSlave(e, msg)
	}
}

func (c *Coordinator) handleAsSlave(e *elevator.Elevator, eMsg message.ElevatorMessage) {
	switch eMsg.EMsgType {
	case message.EMSG_T_StatusReportBroadcast:
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

	case message.EMSG_T_NewToChannel:
		if e.ConnectedToMaster() {
			e.IpRegistery[eMsg.Addr] = eMsg.ID
			e.System.SetStatusReport(eMsg.ID, eMsg.Elevators[eMsg.ID])
		} else if e.Id == eMsg.ID {
			e.SetConnectionState(eMsg)
			e.System.InitializeFromSystemState(eMsg)
		}
	}
}

func (c *Coordinator) handleAsMaster(e *elevator.Elevator, eMsg message.ElevatorMessage) {
	switch eMsg.EMsgType {
	case message.EMSG_T_StatusReport:
		e.SendToCoordinator <- eMsg

	case message.EMSG_T_StatusReportBroadcast:
		e.System.SetStatusReport(eMsg.ID, eMsg.Elevators[eMsg.ID])

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
				taskElevatorId := e.ClosestToTarget(elevs, eMsg.Task)

				if taskElevatorId != "" {
					eMsg.ID = taskElevatorId
					eMsg.BtnStatus = types.Running
				} else {
					eMsg.ID = ""
				}
			}
		}
		eMsg.EMsgType = message.EMSG_T_ButtonPress
		e.SendToCoordinator <- eMsg

	case message.EMSG_T_TaskUpdate:
		e.System.Mutex.Lock()
		e.System.SetRequestStatus(eMsg.ID, eMsg.BtnStatus, eMsg.Task)
		e.System.Mutex.Unlock()
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
		hallRequests, elevs := e.System.Snapshot()
		e.System.Mutex.RUnlock()
		task := e.GetNextTargetFloor(elevs[eMsg.ID], hallRequests)
		if task.Floor != -1 {
			eMsg.Task = task
			eMsg.BtnStatus = types.Running
			e.SendToCoordinator <- eMsg
		}

	case message.EMSG_T_NewToChannel:
		eMsg, id := e.System.RegisterAndSyncElevator(eMsg, e.IpRegistery)
		e.IpRegistery[eMsg.Addr] = id
		e.System.Mutex.Lock()
		e.IsOnline = true
		e.System.Mutex.RUnlock()
		udpAddr, err := udp.StringAddrToUDPAddr(eMsg.Addr)
		if err != nil {
			return
		}
		c.Server.QueueMessage(udpAddr, packet.PROTO_PKT_T_Snapshot, message.ElevatorMessage{})
		c.Server.StartPeerCatchup(udpAddr)

	}
}
