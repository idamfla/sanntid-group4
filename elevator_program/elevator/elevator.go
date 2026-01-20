package elevator

import (
	"fmt"
	"time"

	"elevator_program/utilities"
)

type Elevator struct {
	id int
	currentFloor int
	targetFloor  int
	currentPosition int
	status    ElevatorState

	StatusChan chan utilities.StatusMsg
	TaskChan chan utilities.TaskMsg
}

func (e *Elevator) InitElevator(id int, statusChan chan utilities.StatusMsg, taskChan chan utilities.TaskMsg) {
	e.id = id
	e.currentPosition = 4
	e.targetFloor = 1
	e.currentFloor = 4
	e.status = ElevStateUninitialized

	e.MoveElevator()

	e.targetFloor = 0
	e.status = ElevStateRunning

	e.StatusChan = statusChan
	e.TaskChan = taskChan

	e.StatusChan <-utilities.StatusMsg{e.id, e.currentFloor, e.targetFloor}
}

func (e Elevator) getMotion() utilities.Motion {
	if e.targetFloor == 0 || e.currentPosition == e.targetFloor {
		return utilities.MotionStop
	} else if e.currentFloor < e.targetFloor {
		return utilities.MotionMoveUp
	} else {
		return utilities.MotionMoveDown
	}
}

func (e *Elevator) MoveElevator() {
	ticker := time.NewTicker(50 * time.Millisecond) // faster updates
	defer ticker.Stop()

	for range ticker.C {
		// fmt.Println(e) // TODO: remove db
		switch e.getMotion() {
		case utilities.MotionStop:
			return
		case utilities.MotionMoveUp:
			if e.currentPosition == -1 {
				e.currentPosition = e.currentFloor + 1
				e.currentFloor = e.currentPosition
			} else {
				e.currentFloor = -1
			}
		case utilities.MotionMoveDown:
			if e.currentPosition == -1 {
				e.currentPosition = e.currentFloor - 1
				e.currentFloor = e.currentPosition
			} else {
				e.currentPosition = -1
			}
		}
	}
}

func (e Elevator) String() string {
	return fmt.Sprintf(`Elevator
	id: %d
	current position: %d
	current floor: %d
	target floor: %d
	status: %s`,
		e.id, e.currentFloor, e.currentFloor, e.targetFloor, e.status)
}
