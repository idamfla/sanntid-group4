package main

import (
	"elevator_program/elevator"
	"elevator_program/utilities"
	"fmt"
)

func RunElevatorProgram() {
	statusCh := make(chan utilities.StatusMsg, 10)
	taskCh := make(chan utilities.TaskMsg, 20)

	var tm task_manager.TaskManager
	tm.InitTaskManager()

	var e elevator.Elevator
	
	fmt.Println(e)
	e.InitElevator(1, statusCh, taskCh)
	fmt.Println(e)
}

func main() {
	RunElevatorProgram()
}