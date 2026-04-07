package coordinator

import (
	"elevator_program/elevator"
	"elevator_program/elevio"
	"elevator_program/message"
	"elevator_program/types"
	"fmt"
)

// TODO find out where i change e.Id and e.Ip, should not need mutex but could apear receconditiones

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
		// e.TurnToMaster()         // TODO it is a bit waste to have it here
		c.handleAsMaster(e, msg) // TODO when a new master is elected, it needs to reset the timer on every running task, so no task is lost
	} else {
		// e.TurnToSlave() // TODO have it something like this, just that we only change if e.master is the oposite
		c.handleAsSlave(e, msg)
	}
}

func (c *Coordinator) handleAsSlave(e *elevator.Elevator, eMsg message.ElevatorMessage) {
	switch eMsg.EMsgType {
	case message.EMSG_T_StatusReportBroadcast:
		if eMsg.Addr != e.Ip {
			e.System.SetStatusReport(eMsg.Addr, eMsg.Elevators[eMsg.Addr])
		}

	case message.EMSG_T_TaskUpdate:
		if eMsg.BtnStatus == types.Running {
			// Use SetRequestAsTarget for ALL Running updates (not just own).
			// This ensures the slave reverts the old target to Pending,
			// keeping hall request state in sync with the master.
			e.System.SetRequestAsTarget(eMsg.Addr, eMsg.Task)
		} else {
			e.System.Mutex.Lock()
			e.System.SetRequestStatus(eMsg.Addr, eMsg.BtnStatus, eMsg.Task)
			e.System.Mutex.Unlock()
		}
		// // Don't set cab lamps for other elevators' cab requests
		if !(eMsg.Addr != e.Ip && eMsg.Task.Button == elevio.BT_Cab) {
			e.UpdateBtnLamp(eMsg.Addr, eMsg.BtnStatus, eMsg.Task.Floor, eMsg.Task.Button)
		}

	case message.EMSG_T_SyncSystem: // TODO fix this one to be working for the target and not the target, need to update msgSend as well
		e.System.Mutex.Lock()
		hallCopy := e.System.Intersect(e.System.HallRequests, eMsg.HallRequests)
		e.System.HallRequests = hallCopy
		e.System.Mutex.Unlock()
		e.UpdateMapOfLamps(hallCopy)

		if e.Ip == eMsg.Addr {
			e.System.InitializeFromSystemState(eMsg) // TODO Should i merge with my local cabcalls as well?
		} else {
			e.System.SetStatusReport(eMsg.Addr, eMsg.Elevators[eMsg.Addr])
		}

	case message.EMSG_T_IAmMaster:
		if !e.ConnectedToMaster() { // TODO is this wrong?
			fmt.Println("Lol basenhagen", eMsg)
			e.SetConnectionState(eMsg)
			e.System.Mutex.RLock()
			newMsg := message.ElevatorMessage{
				EMsgType:     message.EMSG_T_IAmMaster,
				ID:           e.Id,
				Addr:         e.Ip,
				HallRequests: e.System.HallRequests,
				Elevators: map[string]types.ElevatorsStatus{
					e.Ip: e.System.Elevators[e.Ip],
				},
			}
			e.System.Mutex.RUnlock()
			e.SendToCoordinator <- newMsg
		}

		e.TurnToSlave()
		var ipAdresses []string
		for ip, elev := range eMsg.Elevators {
			if elev.IsAlive {
				ipAdresses = append(ipAdresses, ip)
			}
		}
		e.System.Mutex.Lock()
		e.UpdateWhoIsALive(ipAdresses)
		e.System.Mutex.Unlock()
	}
}

func (c *Coordinator) handleAsMaster(e *elevator.Elevator, eMsg message.ElevatorMessage) {
	switch eMsg.EMsgType {
	case message.EMSG_T_StatusReport:
		fmt.Println("Halla beeelelelelele \n\n\n\n\n\n\n\n", eMsg)
		e.SendToCoordinator <- eMsg

	case message.EMSG_T_StatusReportBroadcast:
		if eMsg.Addr != e.Ip {
			e.System.SetStatusReport(eMsg.Addr, eMsg.Elevators[eMsg.Addr])
		}

	case message.EMSG_T_ButtonPress:
		if eMsg.BtnStatus != types.NotActive {
			e.System.Mutex.RLock()
			_, elevs := e.System.Snapshot()
			e.System.Mutex.RUnlock()

			if eMsg.Task.Button == elevio.BT_Cab {
				if e.IsNewTargetBetterCab(eMsg.Task, elevs[eMsg.Addr]) {
					eMsg.BtnStatus = types.Running
				} else {
					eMsg.BtnStatus = types.Pending
				}
			} else {
				taskElevatorIp := e.ClosestToTarget(elevs, eMsg.Task)

				if taskElevatorIp != "" {
					eMsg.Addr = taskElevatorIp
					eMsg.BtnStatus = types.Running
				} else {
					eMsg.BtnStatus = types.Pending
				}
			}
		}
		eMsg.EMsgType = message.EMSG_T_ButtonPress
		// e.System.Mutex.RLock()
		// hallCopy := make([][2]types.ButtonStatus, len(e.System.HallRequests)) // TODO don't think we need this anymore
		// copy(hallCopy, e.System.HallRequests)
		// e.System.Mutex.RUnlock()
		// eMsg.HallRequests = hallCopy
		e.SendToCoordinator <- eMsg

	case message.EMSG_T_TaskUpdate:
		if eMsg.BtnStatus == types.Running {
			e.System.SetRequestAsTarget(eMsg.Addr, eMsg.Task)
		} else {
			e.System.Mutex.Lock()
			e.System.SetRequestStatus(eMsg.Addr, eMsg.BtnStatus, eMsg.Task)
			e.System.Mutex.Unlock()
		}
		if !(eMsg.Addr != e.Ip && eMsg.Task.Button == elevio.BT_Cab) {
			e.UpdateBtnLamp(eMsg.Addr, eMsg.BtnStatus, eMsg.Task.Floor, eMsg.Task.Button)
		}

		taskKey := TaskKey{
			Owner:  eMsg.Addr,
			TaskID: eMsg.Task,
		}
		switch eMsg.BtnStatus {
		case types.Running:
			c.TaskMonitor.StartTask(taskKey, e)
		case types.NotActive:
			c.TaskMonitor.FinishTask(taskKey)

			// Assign next Pending task to the elevator that just became free.
			e.System.Mutex.RLock()
			hallRequests, elevs := e.System.Snapshot()
			e.System.Mutex.RUnlock()
			task := e.GetNextTargetFloor(elevs[eMsg.Addr], hallRequests)
			if task.Floor != -1 {
				assignMsg := message.ElevatorMessage{
					EMsgType:  message.EMSG_T_TaskUpdate,
					ID:        e.Id,
					Addr:      eMsg.Addr,
					Task:      task,
					BtnStatus: types.Running,
				}
				e.SendToCoordinator <- assignMsg
			}
		}

	case message.EMSG_T_TaskRequest:
		e.System.Mutex.RLock()
		hallRequests, elevs := e.System.Snapshot()
		// fmt.Println("what is in my system?? \n\n\n\n\n", e.System)
		e.System.Mutex.RUnlock()
		task := e.GetNextTargetFloor(elevs[eMsg.Addr], hallRequests)
		if task.Floor != -1 {
			eMsg.Task = task
			eMsg.BtnStatus = types.Running

			// Need this to clear any ghost running states
			hallcopy := make([][2]types.ButtonStatus, len(hallRequests)) // TODO hallcopy seams kind of unødvendig, cann't i just use hallRequests
			copy(hallcopy, hallRequests)
			eMsg.HallRequests = hallcopy // TODO don't think i need this anymore

			fmt.Println("I have found a new task!! ", eMsg)
			e.SendToCoordinator <- eMsg
		}

	case message.EMSG_T_NewToChannel:
		e.System.Mutex.Lock()
		numFloors := e.NumFloors
		e.IsOnline = true
		e.System.Mutex.Unlock()
		eMsg := e.System.RegisterAndSyncElevator(eMsg, numFloors)

		e.SendToCoordinator <- eMsg

	case message.EMSG_T_SyncedElevator:
		fmt.Println("I am not supposed to do anything here, \n\n\n")

	case message.EMSG_T_IAmMaster:
		c.takeoverAsMaster(e)
		var ipAdresses []string
		for ip, elev := range eMsg.Elevators {
			if elev.IsAlive {
				ipAdresses = append(ipAdresses, ip)
			}
		}
		e.System.Mutex.Lock()
		e.UpdateWhoIsALive(ipAdresses)
		e.System.Mutex.Unlock()

	case message.EMSG_T_SyncSystem:
		e.System.Mutex.Lock()
		hallCopy := e.System.Intersect(e.System.HallRequests, eMsg.HallRequests)
		e.System.HallRequests = hallCopy
		e.System.Mutex.Unlock()
		e.UpdateMapOfLamps(hallCopy)
		e.System.SetStatusReport(eMsg.Addr, eMsg.Elevators[eMsg.Addr])
	}
}

func (c *Coordinator) takeoverAsMaster(e *elevator.Elevator) {
	e.TurnToMaster()
	e.System.Mutex.RLock()
	_, elevs := e.System.Snapshot()
	e.System.Mutex.RUnlock()
	c.TaskMonitor.transferTaskMonitor(elevs, e)
}
