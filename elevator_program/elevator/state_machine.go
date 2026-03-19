package elevator

import (
	"elevator_program/elevio"
	"elevator_program/message"
	"elevator_program/types"
	"fmt"
	"time"
)

// ------------------------
// Motion helper functions
// ------------------------
func (e *Elevator) atTargetFloor(targetFloor int) bool {
	return e.currentFloor == targetFloor && !e.inBetweenFloors && e.currentFloor != -1
}

func (e *Elevator) isTargetValid(targetFloor int) bool {
	return targetFloor >= 0 && targetFloor < e.numFloors
}

func (e *Elevator) getMotion(target int) elevio.MotorDirection {
	if e.atTargetFloor(target) || e.emergencyStop || !e.isTargetValid(target) {
		return elevio.MD_Stop
	} else if e.currentFloor < target {
		return elevio.MD_Up
	} else {
		return elevio.MD_Down
	}
}

func (e *Elevator) updateDirection(target elevio.ButtonEvent, dir elevio.MotorDirection) {
	if dir == elevio.MD_Stop && target.Floor == e.currentFloor {
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
		elevatorCopy := e.System.Elevators[e.Id]
		elevatorCopy.Direction = dir
		e.System.Elevators[e.Id] = elevatorCopy
	}
}

func (e *Elevator) computeNextTargetAndDirection() (elevio.ButtonEvent, elevio.MotorDirection) {
	e.System.Mutex.RLock()
	hallRequests, elevs := e.System.Snapshot()
	e.System.Mutex.RUnlock()
	elevatorStatus := elevs[e.Id]

	nextTarget := e.GetNextTargetFloor(elevatorStatus, hallRequests) // TODO This may be wrong
	if nextTarget.Floor == -1 {
		return elevio.ButtonEvent{Floor: -1}, elevio.MD_Stop
	}

	dir := e.getMotion(nextTarget.Floor)
	return nextTarget, dir
}

func (e *Elevator) uninitializedAction() elevio.MotorDirection {
	if e.currentFloor == -1 {
		return elevio.MD_Down
	}

	if e.currentFloor < e.initFloor {
		return elevio.MD_Up
	}

	if e.currentFloor > e.initFloor {
		return elevio.MD_Down
	}

	return elevio.MD_Stop
}

// ------------------------
// State Machine
// ------------------------
// Updates elevator state when connected to the network
func (e *Elevator) updateElevatorStateOnline() { // TODO rename, this change state and controls the motor
	if e.emergencyStop {
		elevio.SetMotorDirection(elevio.MD_Stop)
		return
	}

	e.System.Mutex.RLock()
	_, elevs := e.System.Snapshot()
	e.System.Mutex.RUnlock()
	elevatorState := elevs[e.Id]
	prevState := elevatorState.State

	// TODO add doorstate switch, e.startTime = time.Now()

	var dir elevio.MotorDirection = elevio.MD_Stop

	if elevatorState.State != types.ES_Uninitialized && e.doorState != DS_Closed {
		elevio.SetMotorDirection(elevio.MD_Stop)
		return
	}

	switch elevatorState.State {
	case types.ES_Uninitialized:
		dir = e.uninitializedAction()

		if dir == elevio.MD_Stop {
			e.clearCurrentFloor(e.currentFloor, elevio.BT_Cab)
			elevatorState.State = types.ES_Idle
			e.doorState = DS_Opening
			// fmt.Println(e)

			e.System.Mutex.RLock()
			_, elevs := e.System.Snapshot()
			e.System.Mutex.RUnlock()

			eMsg := message.ElevatorMessage{
				EMsgType: message.EMSG_T_TaskRequest,
				ID:       e.Id,
				Elevators: map[string]types.ElevatorsStatus{
					e.Id: elevs[e.Id],
				},
			}
			e.SendToCoordinator <- eMsg
		}

	case types.ES_Idle:
		e.System.Mutex.RLock()
		targetFloor := e.System.Elevators[e.Id].Target.Floor
		e.System.Mutex.RUnlock()

		if targetFloor != -1 {
			dir = e.getMotion(targetFloor)
			if dir != elevio.MD_Stop {
				elevatorState.State = types.ES_Moving
			} else {
				e.doorState = DS_Opening
				e.finishedTask(elevatorState.State)
			}
		}

	case types.ES_Moving:
		e.System.Mutex.RLock()
		targetFloor := e.System.Elevators[e.Id].Target.Floor
		e.System.Mutex.RUnlock()

		dir = e.getMotion(targetFloor)

		if dir == elevio.MD_Stop {
			e.doorState = DS_Opening
			elevatorState.State = types.ES_Idle
			e.finishedTask(elevatorState.State)
		}

	case types.ES_EmergencyStop:
		return
	}

	elevio.SetMotorDirection(dir)

	// If state has changed, notify
	e.System.Mutex.Lock()
	if elevatorState.State != e.System.Elevators[e.Id].State {
		elevatorCopy := e.System.Elevators[e.Id]
		elevatorCopy.State = elevatorState.State
		e.System.Elevators[e.Id] = elevatorCopy

		eMsg := message.ElevatorMessage{
			EMsgType: message.EMSG_T_StatusReport,
			ID:       e.Id,
			Elevators: map[string]types.ElevatorsStatus{
				e.Id: e.System.Elevators[e.Id],
			},
		}
		e.System.Mutex.Unlock()
		e.SendToCoordinator <- eMsg
	} else {
		e.System.Mutex.Unlock()
	}

	if prevState != types.ES_Moving && elevatorState.State == types.ES_Moving && dir != elevio.MD_Stop {
		e.markRecoveryVerified()
	}
}

// Updates the elevator state when not connected to the network
func (e *Elevator) updateElevatorStateOffline() { // TODO rename, this change state and controls the motor
	if e.emergencyStop {
		elevio.SetMotorDirection(elevio.MD_Stop)
		return
	}

	e.System.Mutex.RLock()
	_, elevs := e.System.Snapshot()
	elevatorStatus := elevs[e.Id]
	e.System.Mutex.RUnlock()
	prevState := elevatorStatus.State

	// TODO add doorstate switch, e.startTime = time.Now()

	var dir elevio.MotorDirection = elevio.MD_Stop
	var nextTarget elevio.ButtonEvent = elevio.ButtonEvent{Floor: -1, Button: elevio.BT_Cab}

	if elevatorStatus.State != types.ES_Uninitialized && e.doorState != DS_Closed {
		elevio.SetMotorDirection(elevio.MD_Stop)

		return
	}

	switch elevatorStatus.State {
	case types.ES_Uninitialized:
		dir = e.uninitializedAction()

		if dir == elevio.MD_Stop {
			e.clearCurrentFloor(e.currentFloor, elevio.BT_Cab)
			elevatorStatus.State = types.ES_Idle
			e.doorState = DS_Opening
			// fmt.Println(e)
		}

	case types.ES_Idle:
		if elevatorStatus.Target.Floor != -1 {
			dir = e.getMotion(elevatorStatus.Target.Floor)
			if dir != elevio.MD_Stop {
				elevatorStatus.State = types.ES_Moving
			}
		}

		// if elevatorStatus.Target.Floor == -1 {
		// 	nextTarget, dir = e.computeNextTargetAndDirection()
		// 	if nextTarget.Floor != -1 {
		// 		elevatorStatus.Target = nextTarget
		// 	}
		// }
		nextTarget, dir = e.computeNextTargetAndDirection()
		if nextTarget.Floor != -1 {
			elevatorStatus.Target = nextTarget
			e.updateDirection(nextTarget, dir)

			if e.offline {
				e.scheduleRestart = true
			}
		}

		if e.atTargetFloor(elevatorStatus.Target.Floor) {
			e.doorState = DS_Opening
			e.clearCurrentFloor(e.currentFloor, elevatorStatus.Target.Button)
		}

		dir = e.getMotion(elevatorStatus.Target.Floor)
		if dir != elevio.MD_Stop {
			elevatorStatus.State = types.ES_Moving
		}

	case types.ES_Moving:
		dir = e.getMotion(elevatorStatus.Target.Floor)

		if dir == elevio.MD_Stop {
			e.doorState = DS_Opening
			elevatorStatus.State = types.ES_Idle
			e.clearCurrentFloor(e.currentFloor, elevatorStatus.Target.Button)
		}

		// dir = e.getMotion(elevatorStatus.Target.Floor)

		// if dir == elevio.MD_Stop {
		// 	e.doorState = DS_Opening
		// 	elevatorStatus.State = types.ES_Idle
		// } else {
		// 	nextTarget, dir = e.computeNextTargetAndDirection()
		// 	if nextTarget.Floor != -1 && nextTarget != elevatorStatus.Target { // tODO Maybe test that this version still works
		// 		e.System.Mutex.Lock()
		// 		// TODO I don't know if this is the best way to write it but now can use running
		// 		if elevatorStatus.Target.Button == elevio.BT_Cab {
		// 			e.System.Elevators[e.Id].CabRequests[elevatorStatus.Target.Floor] = types.Pending // TODO Need to message that the buttons have changed
		// 		} else {
		// 			e.System.HallRequests[elevatorStatus.Target.Floor][elevatorStatus.Target.Button] = types.Pending // TODO Need to message that the buttons have changed
		// 		}

		// 		if nextTarget.Button == elevio.BT_Cab {
		// 			e.System.Elevators[e.Id].CabRequests[nextTarget.Floor] = types.Running
		// 		} else {
		// 			e.System.HallRequests[nextTarget.Floor][nextTarget.Button] = types.Running
		// 		}
		// 		e.System.Mutex.Unlock()
		// 		elevatorStatus.Target = nextTarget
		// 		e.updateDirection(nextTarget, dir)
		// 	}
		// }
	case types.ES_EmergencyStop:
		elevio.SetMotorDirection(elevio.MD_Stop)

		return
	}

	if prevState != types.ES_Moving && elevatorStatus.State == types.ES_Moving && dir != elevio.MD_Stop {
		e.markRecoveryVerified()
	}
	elevio.SetMotorDirection(dir)

	e.checkOfflineRestart()

	e.System.Elevators[e.Id] = elevatorStatus
	e.System.Mutex.Lock()
	elevatorCopy := e.System.Elevators[e.Id]
	elevatorCopy.State = elevatorStatus.State
	e.System.Elevators[e.Id] = elevatorCopy
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
			if e.IsOnline {
				e.updateElevatorStateOnline()
			} else {
				e.updateElevatorStateOffline()
			}
		}
	}
}

// func (e *Elevator) finishedTask(state types.ElevatorState) {
// 	e.System.Mutex.Lock()
// 	defer e.System.Mutex.Unlock()

// 	target := e.System.Elevators[e.Id].Target
// 	if target.Floor == -1 {
// 		return
// 	}

// 	eMsg := message.ElevatorMessage{
// 		EMsgType:  message.EMSG_T_ButtonPress,
// 		ID:        e.Id,
// 		Task:      target,
// 		BtnStatus: types.NotActive,
// 	}
// 	e.SendToCoordinator <- eMsg

// 	elevatorCopy := e.System.Elevators[e.Id]
// 	elevatorCopy.State = state
// 	elevatorCopy.CabRequests[target.Floor] = types.NotActive
// 	elevatorCopy.Target = elevio.ButtonEvent{Floor: -1, Button: elevio.BT_HallUp} // TODO this might screw me over
// 	e.System.Elevators[e.Id] = elevatorCopy

// 	eMsg.EMsgType = message.EMSG_T_TaskRequest
// 	eMsg.Elevators = map[string]types.ElevatorsStatus{
// 		e.Id: e.System.Elevators[e.Id],
// 	}
// 	e.SendToCoordinator <- eMsg
// }

func (e *Elevator) finishedTask(state types.ElevatorState) { // TODO as claude how to fix dependancy bug
	e.System.Mutex.Lock()

	target := e.System.Elevators[e.Id].Target
	if target.Floor == -1 {
		e.System.Mutex.Unlock()
		return
	}

	hallRequests, elevs := e.System.Snapshot()

	elevatorCopy := elevs[e.Id]
	elevatorCopy.State = state
	elevatorCopy.Target = elevio.ButtonEvent{Floor: -1, Button: elevio.BT_HallUp}
	elevs[e.Id] = elevatorCopy
	e.System.Elevators[e.Id] = elevatorCopy

	e.System.Mutex.Unlock()

	finishedMsg := message.ElevatorMessage{
		EMsgType:  message.EMSG_T_ButtonPress,
		ID:        e.Id,
		Task:      target,
		BtnStatus: types.NotActive,
	}

	e.SendToCoordinator <- finishedMsg

	if e.IsMaster {
		task := e.GetNextTargetFloor(elevs[e.Id], hallRequests)

		if task.Floor != -1 {
			assignMsg := message.ElevatorMessage{
				EMsgType:  message.EMSG_T_TaskUpdate,
				ID:        e.Id,
				Task:      task,
				BtnStatus: types.Running,
			}
			e.SendToCoordinator <- assignMsg
		}

	} else {
		requestMsg := message.ElevatorMessage{
			EMsgType: message.EMSG_T_TaskRequest,
			ID:       e.Id,
			Task:     target,
			Elevators: map[string]types.ElevatorsStatus{
				e.Id: elevatorCopy,
			},
		}

		e.SendToCoordinator <- requestMsg
	}
}
