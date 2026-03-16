package elevator

import (
	"elevator_program/elevio"
	"elevator_program/message"
	"elevator_program/system"
	"elevator_program/types"
	"time"
)

// TODO This function is wierd, either we need to have it as e or something else if it is msg sending
func (e *Elevator) HandleLostConnection(senderId string) {
	if senderId == "" || time.Since(e.lostComsTimer) > 4*time.Second {
		// Need to schedule a restart
		// TODO It is fault tolerance that should take the time maybe
	} else {
		e.ackCounterLostComs++
		if e.ackCounterLostComs >= len(e.System.Elevators)-1 { // TODO something not right here, maybe master and a slave has died, then you are waiting for the dead slave to respond as well
			// Need to reset the timer
			// Need to start election
		}
	}
}

func (e *Elevator) ConnectedToMaster() bool {
	return e.connectedToMaster
}

func (e Elevator) UpdateBtnLamp(btnStatus types.ButtonStatus, floor int, button elevio.ButtonType) {
	if btnStatus == types.NotActive {
		e.clearHallLamp(floor, button)
	} else {
		elevio.SetButtonLamp(button, floor, true)
	}
}

func (e *Elevator) SetConnectionState(msg message.Message) {
	e.Id = msg.Id
	e.IsMaster = false
	e.connectedToMaster = true
	e.IsOnline = true
	// e.elevatorState = types.ES_Idle // TODO Do I need this one here?
	for id, elevator := range e.System.Elevators {
		e.IpRegistery[elevator.Ip] = id
	}
}

// TODO Probably don't need, just for testing
func (e *Elevator) ClearElevator(numFloors int) {
	e.System.HallRequests = make([][2]types.ButtonStatus, numFloors)
	e.System.Elevators = make(map[string]types.ElevatorsStatus)
	e.nextTarget = elevio.ButtonEvent{
		Floor:  -1,
		Button: elevio.BT_HallUp,
	}
	elevio.SetMotorDirection(0)
	// TODO if i want to test this one, have to change to systemstate
	// e.elevatorState = types.ES_EmergencyStop
}

// TODO Don't need
func (e Elevator) Create_slave(system system.System) Elevator {
	slave := Elevator{
		Id:       "2",
		IsMaster: false,
		System:   system,
	}
	return slave
}

func (e *Elevator) SetRequestAsTarget(task elevio.ButtonEvent) {
	// TODO I think it is wierd that I call system from here. The whole purpose of this was to seperate sytsem and elevator
	if e.nextTarget.Floor != -1 {
		e.System.SetRequestStatus(e.Id, types.Pending, e.nextTarget)
	}

	e.System.SetRequestStatus(e.Id, types.Running, task)

	e.nextTarget = task

	if task.Floor > e.currentFloor {
		e.direction = elevio.MD_Up
	} else if task.Floor < e.currentFloor {
		e.direction = elevio.MD_Down
	}
}

func (e *Elevator) TurnToMaster() {
	e.IsOnline = true
	e.IsMaster = true
}
