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
			elevio.SetButtonLamp(ev.Button, ev.Floor, true) // TODO don't turn on lamp before master says to do so
		case EV_FloorSensor:
			if ev.Floor != -1 { e.currentFloor = ev.Floor}
		case EV_Obstruction:
			// TODO if door_open, and obsturcion pressed, e.state = obstruction
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
			e.ClearFloor(e.currentFloor) // TODO don't remove floor and turn off button in a function that says it only cares about the motor ... 
			e.ClearButtonLamp(e.currentFloor)

			e.state = ES_Idle
		}

	case ES_Idle:
		nextTarget := e.UpdateTargetFloor() // sets targetFloor, TODO does this need to be a function or can i do it directly
		dir := e.GetMotion()
		if dir != elevio.MD_Stop {
			e.state = ES_Moving
			elevio.SetMotorDirection(dir)
		}

		if e.currentFloor != -1 && e.currentFloor == e.targetFloor { // TODO is it here bc if someone spams the button on the floor you're at?
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
			e.state = ES_Idle // TODO state to doorOpen
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
		case ev := <-e.eventsCh:
			e.handleEvent(ev)
		case <-ticker.C:
			e.updateMotor()
		}
		fmt.Println(e) // DB
	}
}
