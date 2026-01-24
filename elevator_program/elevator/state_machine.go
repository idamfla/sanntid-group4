package elevator

import (
	"elevator_program/elevio"
	"fmt"
	"time"
)

func (e *Elevator) handleEvent(ev ElevatorEvent) {
	switch ev.Type {
		case EV_EmergencyStop:
			elevio.SetMotorDirection(elevio.MD_Stop)
			elevio.SetStopLamp(ev.EmergencyStop)
			e.emergencyStop = ev.EmergencyStop
		case EV_ButtonPress:
			e.floorRequests[ev.Floor][ev.Button] = true
			elevio.SetButtonLamp(ev.Button, ev.Floor, true)
		case EV_FloorSensor:
			if ev.Floor != -1 { e.currentFloor = ev.Floor}
	}
}

func (e *Elevator) updateMotor() {
	if e.emergencyStop { return }

	switch e.state {
	case ES_Uninitialized:
		if e.currentFloor == -1 {
			elevio.SetMotorDirection(elevio.MD_Down) // always move down until a floor sensor triggers
			return
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
	}
}

func (e *Elevator) ElevatorStateMachine() {
	ticker := time.NewTicker(50 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case ev := <-e.stateMachineCh:
			e.handleEvent(ev)
		case <-ticker.C:
			e.updateMotor()
		}
		fmt.Println(e) // DB
	}
}
