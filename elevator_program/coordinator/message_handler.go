package coordinator

import (
	"elevator_program/elevator"
	"elevator_program/elevio"
	"elevator_program/message"
	"elevator_program/types"
	"fmt"
)

func (c *Coordinator) MessageListener(e *elevator.Elevator) {
	fmt.Println("MESSAGE LISTENER STARTED", e.Id)
	for ePkt := range c.msgRecieveCh {
		fmt.Println(e)
		eMsg := ePkt.EMsg
		fmt.Println("thththt \n\n\n\n\n", eMsg)
		c.MessageHandler(e, eMsg)
		if ePkt.Done != nil {
			func() {
				defer func() { recover() }()
				ePkt.Done <- struct{}{}
			}()
		}
	}
}

func (c *Coordinator) MessageHandler(e *elevator.Elevator, msg message.ElevatorMessage) {
	if c.Server.IsMaster() {
		e.TurnToMaster() // TODO it is a bit waste to have it here
		c.handleAsMaster(e, msg)
	} else {
		c.handleAsSlave(e, msg)
	}
}

func (c *Coordinator) handleAsSlave(e *elevator.Elevator, eMsg message.ElevatorMessage) {
	switch eMsg.EMsgType {
	case message.EMSG_T_StatusReportBroadcast:
		e.IpRegistery[eMsg.Addr] = eMsg.ID // TODO Do i need the ipRegistery, I use it in registerAndSyncElev.
		e.System.SetStatusReport(eMsg.ID, eMsg.Elevators[eMsg.ID])
		e.System.Mutex.RLock()
		fmt.Println("Do i get fucked here?? \n\n\n\n\n\n\n", e.System)
		e.System.Mutex.RUnlock()

	case message.EMSG_T_TaskUpdate:
		if e.Id == eMsg.ID && eMsg.BtnStatus == types.Running {
			e.System.SetRequestAsTarget(eMsg.ID, eMsg.Task)
		} else {
			e.System.Mutex.Lock()
			e.System.SetRequestStatus(eMsg.ID, eMsg.BtnStatus, eMsg.Task)
			e.System.Mutex.Unlock()
		}
		e.UpdateBtnLamp(eMsg.BtnStatus, eMsg.Task.Floor, eMsg.Task.Button)

		fmt.Println("yooooooo \n\n\n\n\n\n\n", eMsg.BtnStatus, eMsg.ID)

		e.System.Mutex.RLock()
		fmt.Println("What is my id and ip??", e.Id, e.Ip) // TODO I need to change so everything analyses ip/addr instead of id,
		e.System.Mutex.RUnlock()

		if eMsg.BtnStatus == types.NotActive && e.Ip == eMsg.ID { // TODO I think this may prevent raceconditions but something is wrong
			fmt.Println("THISISIS")
			requestMsg := message.ElevatorMessage{
				EMsgType: message.EMSG_T_TaskRequest,
				ID:       e.Id,
				Task:     eMsg.Task,
			}
			e.SendToCoordinator <- requestMsg
		}

	case message.EMSG_T_NewToChannel:
		// if e.ConnectedToMaster() {
		// 	e.IpRegistery[eMsg.Addr] = eMsg.ID
		// 	e.System.SetStatusReport(eMsg.ID, eMsg.Elevators[eMsg.ID])
		// } else if e.Id == eMsg.ID {
		e.SetConnectionState(eMsg)
		e.System.InitializeFromSystemState(eMsg)
		e.System.Mutex.RLock()
		fmt.Println("I have synced !!!! \n\n\n\n\n\n", e.System)
		e.System.Mutex.RUnlock()
		// }
	}
}

func (c *Coordinator) handleAsMaster(e *elevator.Elevator, eMsg message.ElevatorMessage) {
	switch eMsg.EMsgType {
	case message.EMSG_T_StatusReport:
		fmt.Println("Halla beeelelelelele \n\n\n\n\n\n\n\n", eMsg)
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
		if eMsg.BtnStatus == types.Running {
			e.System.SetRequestAsTarget(eMsg.ID, eMsg.Task)
		} else {
			e.System.Mutex.Lock()
			e.System.SetRequestStatus(eMsg.ID, eMsg.BtnStatus, eMsg.Task)
			e.System.Mutex.Unlock()
		}
		if !(eMsg.ID != e.Id && eMsg.Task.Button == elevio.BT_Cab) {
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
		fmt.Println("what is in my system?? \n\n\n\n\n", e.System)
		e.System.Mutex.RUnlock()
		task := e.GetNextTargetFloor(elevs[eMsg.ID], hallRequests)
		if task.Floor != -1 {
			eMsg.Task = task
			eMsg.BtnStatus = types.Running
			fmt.Println("I have found a new task!! ", eMsg)
			e.SendToCoordinator <- eMsg
		}

	case message.EMSG_T_NewToChannel:
		e.System.Mutex.Lock()
		numFloors := e.NumFloors
		e.IsOnline = true
		e.System.Mutex.Unlock()
		eMsg, id := e.System.RegisterAndSyncElevator(eMsg, e.IpRegistery, numFloors)
		e.IpRegistery[eMsg.Addr] = id

		fmt.Println("Before syncing \n\n\n\n\n", eMsg)
		e.System.Mutex.RLock()
		fmt.Println("Have I updated myself?? \n\n", e.System)
		e.System.Mutex.RUnlock()
		e.SendToCoordinator <- eMsg

	case message.EMSG_T_SyncedElevator:
		fmt.Println("I am not supposed to do anything here, i am a stupid bird \n\n\n")
	}
}
