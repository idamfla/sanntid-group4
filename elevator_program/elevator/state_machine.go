package elevator

import (
	"elevator_program/elevio"
	"elevator_program/message"
	"elevator_program/types"
	"fmt"
	"time"
)

func (e *Elevator) atTargetFloor(targetFloor int) bool {
	e.mu.Lock()
	floor := e.currentFloor
	between := e.inBetweenFloors
	e.mu.Unlock()
	return floor == targetFloor && !between && floor != -1
}

func (e *Elevator) isTargetValid(targetFloor int) bool {
	return targetFloor >= 0 && targetFloor < e.NumFloors
}

func (e *Elevator) getMotion(target int) elevio.MotorDirection {
	e.mu.Lock()
	estop := e.emergencyStop
	floor := e.currentFloor
	e.mu.Unlock()
	if e.atTargetFloor(target) || estop || !e.isTargetValid(target) {
		return elevio.MD_Stop
	} else if floor < target {
		return elevio.MD_Up
	} else {
		return elevio.MD_Down
	}
}

func (e *Elevator) updateDirection(target elevio.ButtonEvent, dir elevio.MotorDirection) {
	e.mu.Lock()
	floor := e.currentFloor
	e.mu.Unlock()
	if dir == elevio.MD_Stop && target.Floor == floor {
		switch target.Button {
		case elevio.BT_HallUp:
			dir = elevio.MD_Up
		case elevio.BT_HallDown:
			dir = elevio.MD_Down
		}
	}

	if dir != elevio.MD_Stop {
		e.System.Mutex.Lock()
		defer e.System.Mutex.Unlock()
		elevatorCopy := e.System.Elevators[e.Ip]
		elevatorCopy.Direction = dir
		e.System.Elevators[e.Ip] = elevatorCopy
	}
}

func (e *Elevator) computeNextTargetAndDirection() (elevio.ButtonEvent, elevio.MotorDirection) {
	e.System.Mutex.RLock()
	hallRequests, elevs := e.System.Snapshot()
	e.System.Mutex.RUnlock()
	elevatorStatus := elevs[e.Ip]

	nextTarget := e.GetNextTargetFloor(elevatorStatus, hallRequests)
	if nextTarget.Floor == -1 {
		return elevio.ButtonEvent{Floor: -1}, elevio.MD_Stop
	}

	dir := e.getMotion(nextTarget.Floor)
	return nextTarget, dir
}

func (e *Elevator) uninitializedAction() elevio.MotorDirection {
	e.mu.Lock()
	floor := e.currentFloor
	e.mu.Unlock()

	if floor == -1 {
		return elevio.MD_Down
	}
	if floor < e.initFloor {
		return elevio.MD_Up
	}
	if floor > e.initFloor {
		return elevio.MD_Down
	}
	return elevio.MD_Stop
}

func (e *Elevator) updateElevatorStateOnline() {
	e.mu.Lock()
	estop := e.emergencyStop
	ds := e.doorState
	e.mu.Unlock()

	if estop {
		elevio.SetMotorDirection(elevio.MD_Stop)
		e.System.Mutex.Lock()
		elevatorCopy := e.System.Elevators[e.Ip]
		elevatorCopy.State = types.ES_EmergencyStop
		e.System.Elevators[e.Ip] = elevatorCopy
		e.System.Mutex.Unlock()
		return
	}

	e.System.Mutex.RLock()
	_, elevs := e.System.Snapshot()
	e.System.Mutex.RUnlock()
	elevatorState := elevs[e.Ip]

	var dir elevio.MotorDirection = elevio.MD_Stop

	if elevatorState.State != types.ES_Uninitialized && ds != DS_Closed {
		elevio.SetMotorDirection(elevio.MD_Stop)
		return
	}

	switch elevatorState.State {
	case types.ES_Uninitialized:
		dir = e.uninitializedAction()

		if dir == elevio.MD_Stop {
			e.clearCurrentFloor(e.currentFloor, elevio.BT_Cab)
			elevatorState.State = types.ES_Idle
			e.mu.Lock()
			e.doorState = DS_Opening
			e.mu.Unlock()
			fmt.Println(e)

			e.System.Mutex.RLock() // TODO i don't think we need this one
			_, elevs := e.System.Snapshot()
			e.System.Mutex.RUnlock()

			eMsg := message.ElevatorMessage{
				EMsgType: message.EMSG_T_TaskRequest, // TODO i don't know if we should delete this one
				ID:       e.Id,
				Addr:     e.Ip,
				Elevators: map[string]types.ElevatorsStatus{
					e.Ip: elevs[e.Ip],
				},
			}
			e.SendToCoordinator <- eMsg
		}

	case types.ES_Idle:
		e.System.Mutex.RLock()
		targetFloor := e.System.Elevators[e.Ip].Target.Floor
		e.System.Mutex.RUnlock()

		if targetFloor != -1 {
			dir = e.getMotion(targetFloor)
			if dir != elevio.MD_Stop {
				elevatorState.State = types.ES_Moving
			} else {
				e.mu.Lock()
				if e.doorState == DS_Closed {
					e.doorState = DS_Opening
				}
				e.mu.Unlock()
				e.finishedTask(elevatorState.State)
			}
		}

	case types.ES_Moving:
		e.System.Mutex.RLock()
		targetFloor := e.System.Elevators[e.Ip].Target.Floor
		e.System.Mutex.RUnlock()

		dir = e.getMotion(targetFloor)

		if dir == elevio.MD_Stop {
			e.mu.Lock()
			e.doorState = DS_Opening
			e.mu.Unlock()
			elevatorState.State = types.ES_Idle
			e.finishedTask(elevatorState.State)
		}

	case types.ES_EmergencyStop:
		return
	}

	elevio.SetMotorDirection(dir)

	e.System.Mutex.Lock()
	if elevatorState.State != e.System.Elevators[e.Ip].State {
		elevatorCopy := e.System.Elevators[e.Ip]
		elevatorCopy.State = elevatorState.State
		e.System.Elevators[e.Ip] = elevatorCopy

		eMsg := message.ElevatorMessage{
			EMsgType: message.EMSG_T_StatusReport,
			ID:       e.Id,
			Addr:     e.Ip,
			Elevators: map[string]types.ElevatorsStatus{
				e.Ip: elevatorCopy,
			},
		}
		e.System.Mutex.Unlock()
		e.SendToCoordinator <- eMsg
	} else {
		e.System.Mutex.Unlock()
	}

	if e.shouldMarkRecoveryVerified() {
		e.markRecoveryVerified()
	}
}

func (e *Elevator) updateElevatorStateOffline() {
	e.mu.Lock()
	estop := e.emergencyStop
	ds := e.doorState
	e.mu.Unlock()

	if estop {
		elevio.SetMotorDirection(elevio.MD_Stop)
		return
	}

	e.System.Mutex.RLock()
	_, elevs := e.System.Snapshot()
	elevatorStatus := elevs[e.Ip]
	e.System.Mutex.RUnlock()

	var dir elevio.MotorDirection = elevio.MD_Stop
	var nextTarget elevio.ButtonEvent = elevio.ButtonEvent{Floor: -1, Button: elevio.BT_Cab}

	if elevatorStatus.State != types.ES_Uninitialized && ds != DS_Closed {
		elevio.SetMotorDirection(elevio.MD_Stop)
		return
	}

	switch elevatorStatus.State {
	case types.ES_Uninitialized:
		dir = e.uninitializedAction()

		if dir == elevio.MD_Stop {
			e.clearCurrentFloor(e.currentFloor, elevio.BT_Cab)
			elevatorStatus.State = types.ES_Idle
			e.mu.Lock()
			e.doorState = DS_Opening
			e.mu.Unlock()
			fmt.Println(e)
		}

	case types.ES_Idle:
		if elevatorStatus.Target.Floor != -1 {
			dir = e.getMotion(elevatorStatus.Target.Floor)
			if dir != elevio.MD_Stop {
				elevatorStatus.State = types.ES_Moving
			}
		}

		nextTarget, dir = e.computeNextTargetAndDirection()
		if nextTarget.Floor != -1 {
			elevatorStatus.Target = nextTarget
			e.updateDirection(nextTarget, dir)
			e.System.Mutex.Lock()
			if nextTarget.Button == elevio.BT_Cab {
				elevatorStatus.CabRequests[nextTarget.Floor] = types.Running
			} else {
				e.System.HallRequests[nextTarget.Floor][nextTarget.Button] = types.Running
			}
			e.System.Mutex.Unlock()

			e.mu.Lock()
			if !e.IsOnline {
				e.scheduleRestart = true
			}
			e.mu.Unlock()
		}

		if e.atTargetFloor(elevatorStatus.Target.Floor) {
			e.mu.Lock()
			e.doorState = DS_Opening
			e.mu.Unlock()
			e.clearCurrentFloor(e.currentFloor, elevatorStatus.Target.Button)
		}

		dir = e.getMotion(elevatorStatus.Target.Floor)
		if dir != elevio.MD_Stop {
			elevatorStatus.State = types.ES_Moving
		}

	case types.ES_Moving:
		dir = e.getMotion(elevatorStatus.Target.Floor)

		if dir == elevio.MD_Stop {
			e.mu.Lock()
			e.doorState = DS_Opening
			e.mu.Unlock()
			elevatorStatus.State = types.ES_Idle
			e.clearCurrentFloor(e.currentFloor, elevatorStatus.Target.Button)
		}

	case types.ES_EmergencyStop:
		elevio.SetMotorDirection(elevio.MD_Stop)

		return
	}

	if e.shouldMarkRecoveryVerified() {
		e.markRecoveryVerified()
	}
	elevio.SetMotorDirection(dir)

	e.checkOfflineRestart()

	e.System.Mutex.Lock()
	e.System.Elevators[e.Ip] = elevatorStatus
	e.System.Mutex.Unlock()
}

func (e *Elevator) RunElevatorStateMachine() {

	defer e.wg.Done()
	fmt.Println("ELEVATOR STATE MACHINE STARTED")
	ticker := time.NewTicker(50 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-e.stop:
			return
		case <-ticker.C:
			e.mu.Lock()
			online := e.IsOnline
			e.mu.Unlock()
			if online {
				e.updateElevatorStateOnline()
			} else {
				e.updateElevatorStateOffline()
			}
		}
	}
}

func (e *Elevator) finishedTask(state types.ElevatorState) {
	e.System.Mutex.Lock()

	target := e.System.Elevators[e.Ip].Target
	if target.Floor == -1 {
		e.System.Mutex.Unlock()
		return
	}

	_, elevs := e.System.Snapshot()

	elevatorCopy := elevs[e.Ip]
	// elevatorCopy.State = state
	elevatorCopy.Target = elevio.ButtonEvent{Floor: -1, Button: elevio.BT_HallUp}
	elevs[e.Ip] = elevatorCopy
	e.System.Elevators[e.Ip] = elevatorCopy

	e.System.Mutex.Unlock()

	finishedMsg := message.ElevatorMessage{
		EMsgType:  message.EMSG_T_ButtonPress,
		ID:        e.Id,
		Addr:      e.Ip,
		Task:      target,
		BtnStatus: types.NotActive,
	}

	e.SendToCoordinator <- finishedMsg

	// e.mu.Lock()
	// isMaster := e.IsMaster
	// e.mu.Unlock()
	// if isMaster {
	// 	task := e.GetNextTargetFloor(elevs[e.Ip], hallRequests)

	// 	if task.Floor != -1 {
	// 		assignMsg := message.ElevatorMessage{ // TODO this may cause errors
	// 			EMsgType:  message.EMSG_T_TaskUpdate,
	// 			ID:        e.Id,
	// 			Addr:      e.Ip,
	// 			Task:      task,
	// 			BtnStatus: types.Running,
	// 		}
	// 		e.SendToCoordinator <- assignMsg
	// 	}

	// }
}
