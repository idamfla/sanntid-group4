package elevator

import (
	"elevator_program/elevio"
	"fmt"
)

func (e *Elevator) ElevatorStateMachine() {
	for {
		var ev ElevatorEvent
		select {
		case ev = <-e.eventsCh:
			switch ev.Type {
			case EV_FloorSensor:
				if ev.Floor != -1 {
					e.currentFloor = ev.Floor
				}
			case EV_ButtonPress:
				e.floorRequests[ev.Floor][ev.Button] = true
				elevio.SetButtonLamp(ev.Button, ev.Floor, true)
				// TODO send press to master
			case EV_Obstruction:
				if ev.Obstruction {
					e.state = ES_Obstruction
				}
			case EV_EmergencyStop:
				if ev.EmergencyStop {
					elevio.SetStopLamp(true)
					e.state = ES_EmergencyStop
					continue
				} else {
					elevio.SetStopLamp(false)
					e.state = ES_Idle
				}
			case EV_TaskAssigned:
				continue
			case EV_TaskCompleted:
				continue
			}
		}

		switch e.state {
		case ES_Uninitialized:
			if e.currentFloor == -1 {
				elevio.SetMotorDirection(elevio.MD_Down) // always move down until a floor sensor triggers
				continue
			}

			if e.currentFloor < e.initFloor {
				elevio.SetMotorDirection(elevio.MD_Up)
			} else if e.currentFloor > e.initFloor {
				elevio.SetMotorDirection(elevio.MD_Down)
			} else {
				elevio.SetMotorDirection(elevio.MD_Stop)
				e.ClearFloor(e.currentFloor)
				e.ClearButtonLamp(e.currentFloor)

				e.state = ES_Idle
			}

		case ES_Idle:
			nextTarget := e.UpdateTargetFloor() // sets targetFloor
			dir := e.GetMotion()
			if dir != elevio.MD_Stop {
				e.state = ES_Moving
				elevio.SetMotorDirection(dir)
			}

			if e.currentFloor != -1 && e.currentFloor == e.targetFloor {
				elevio.SetMotorDirection(elevio.MD_Stop)
				e.ClearFloor(e.currentFloor)
				e.ClearButtonLamp(e.currentFloor)
			}
			e.UpdateLastDirection(nextTarget, dir)

		case ES_Moving:
			if e.currentFloor == e.targetFloor {
				elevio.SetMotorDirection(elevio.MD_Stop)
				e.ClearFloor(e.currentFloor)
				e.ClearButtonLamp(e.currentFloor)
				e.state = ES_Idle
			} else {
				nextTarget := e.UpdateTargetFloor()
				dir := e.GetMotion()
				elevio.SetMotorDirection(dir)
				e.UpdateLastDirection(nextTarget, dir)
			}
		case ES_EmergencyStop:
			elevio.SetMotorDirection(elevio.MD_Stop)
			if !ev.EmergencyStop {
				elevio.SetStopLamp(false)
				e.state = ES_Idle
			}
		}
		fmt.Println(e) // DB
	}
}
