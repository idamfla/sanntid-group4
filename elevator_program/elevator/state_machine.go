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
func (e Elevator) atTargetFloor() bool {
	return e.currentFloor == e.System.Elevators[e.Id].Target.Floor && !e.inBetweenFloors && e.currentFloor != -1
}

func (e Elevator) isTargetValid() bool {
	return e.System.Elevators[e.Id].Target.Floor >= 0 && e.System.Elevators[e.Id].Target.Floor < len(e.hallRequests)
}

func (e Elevator) getMotion(target int) elevio.MotorDirection {
	if e.atTargetFloor() || e.emergencyStop || !e.isTargetValid() {
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
		elevatorCopy := e.System.Elevators[e.Id]
		elevatorCopy.Direction = dir
		e.System.Elevators[e.Id] = elevatorCopy
	}
}

func (e Elevator) computeNextTargetAndDirection() (elevio.ButtonEvent, elevio.MotorDirection) {
	nextTarget := getNextTargetFloor(e)
	if nextTarget.Floor == -1 {
		return elevio.ButtonEvent{Floor: -1}, elevio.MD_Stop
	}

	dir := e.getMotion(nextTarget.Floor)
	return nextTarget, dir
}

func (e Elevator) uninitializedAction() elevio.MotorDirection {
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

	elevatorState := e.System.Elevators[e.Id]

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

			msg := message.Message{
				MsgType: types.MSG_T_TaskRequest,
				Id:      e.Id,
				Elevators: map[string]types.ElevatorsStatus{
					e.Id: e.System.Elevators[e.Id],
				},
			}
			e.SendToProtocol <- msg
		}

	case types.ES_Idle:
		if e.System.Elevators[e.Id].Target.Floor != -1 {
			dir = e.getMotion(e.System.Elevators[e.Id].Target.Floor)
			if dir != elevio.MD_Stop {
				elevatorState.State = types.ES_Moving
			}
		}

	case types.ES_Moving:
		dir = e.getMotion(e.System.Elevators[e.Id].Target.Floor)

		if dir == elevio.MD_Stop {
			e.doorState = DS_Opening
			elevatorState.State = types.ES_Idle
			msg := message.Message{
				MsgType:   types.MSG_T_ButtonPress,
				Id:        e.Id,
				Task:      e.System.Elevators[e.Id].Target,
				BtnStatus: types.NotActive,
			}
			e.SendToProtocol <- msg

			// TODO We should clean this up
			elevatorCopy := e.System.Elevators[e.Id]
			elevatorCopy.State = elevatorState.State
			e.System.Elevators[e.Id] = elevatorCopy

			msg.MsgType = types.MSG_T_TaskRequest
			msg.Elevators = map[string]types.ElevatorsStatus{
				e.Id: e.System.Elevators[e.Id],
			}
			e.SendToProtocol <- msg
		}

	case types.ES_EmergencyStop:
		return
	}
	elevatorCopy := e.System.Elevators[e.Id]
	elevatorCopy.State = elevatorState.State
	e.System.Elevators[e.Id] = elevatorCopy

	// If state has changed, notify
	if elevatorState.State != e.System.Elevators[e.Id].State {
		msg := message.Message{
			MsgType: types.MSG_T_StatusReport,
			Id:      e.Id,
			Elevators: map[string]types.ElevatorsStatus{
				e.Id: e.System.Elevators[e.Id],
			},
		}
		e.SendToProtocol <- msg
	}

	elevio.SetMotorDirection(dir)
}

// Updates the elevator state when not connected to the network
func (e *Elevator) updateElevatorStateOffline() { // TODO rename, this change state and controls the motor
	if e.emergencyStop {
		elevio.SetMotorDirection(elevio.MD_Stop)
		return
	}

	elevatorStatus := e.System.Elevators[e.Id]

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

		nextTarget, dir = e.computeNextTargetAndDirection()
		if nextTarget.Floor != -1 {
			elevatorStatus.Target = nextTarget
			e.updateDirection(nextTarget, dir)
		}

		if e.atTargetFloor() { // TODO is it here bc if someone spams the button on the floor you're at?
			// e.doorState = open
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

		dir = e.getMotion(elevatorStatus.Target.Floor)

		if dir == elevio.MD_Stop {
			e.doorState = DS_Opening
			elevatorStatus.State = types.ES_Idle
		} else {
			nextTarget, dir = e.computeNextTargetAndDirection()
			if nextTarget.Floor != -1 { // tODO Maybe test that this version still works
				// TODO I don't know if this is the best way to write it but now can use running
				if elevatorStatus.Target.Button == elevio.BT_Cab {
					e.System.Elevators[e.Id].CabRequests[elevatorStatus.Target.Floor] = types.Pending // TODO Need to message that the buttons have changed
				} else {
					e.System.HallRequests[elevatorStatus.Target.Floor][elevatorStatus.Target.Button] = types.Pending // TODO Need to message that the buttons have changed
				}

				if nextTarget.Button == elevio.BT_Cab {
					e.System.Elevators[e.Id].CabRequests[nextTarget.Floor] = types.Running
				} else {
					e.System.HallRequests[nextTarget.Floor][nextTarget.Button] = types.Running
				}
				elevatorStatus.Target = nextTarget
				e.updateDirection(nextTarget, dir)
			}
		}
	case types.ES_EmergencyStop:
		return
	}

	elevio.SetMotorDirection(dir)

	e.System.Elevators[e.Id] = elevatorStatus

	msg := message.Message{
		MsgType: types.MSG_T_NewToChannel,
		Id:      e.Id,
		Ip:      e.Ip,
		Elevators: map[string]types.ElevatorsStatus{
			e.Id: e.System.Elevators[e.Id],
		},
	}
	fmt.Println("Trying to send to network, ", e.Id)
	var lastSend time.Time

	if time.Since(lastSend) > 200*time.Millisecond {
		e.SendToProtocol <- msg
		lastSend = time.Now()
	}
}

func (e *Elevator) RunElevatorStateMachine() {
	fmt.Println("ELEVATOR STATE MACHINE STARTED")
	ticker := time.NewTicker(50 * time.Millisecond)
	defer ticker.Stop()

	for range ticker.C {
		if e.IsOnline {
			e.updateElevatorStateOnline()
		} else {
			e.updateElevatorStateOffline()
		}
	}
}
