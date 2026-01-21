package main

import (
	"elevator_program/elevator"
	// "elevator_program/utilities"
	// "elevator_program/elevio"
	"fmt"
)

func testElevator() {
	var e elevator.Elevator
	
	id := 1
	numFloors := 4
	initTargetFloor := 0

	fmt.Println(e)
	// e.InitElevator(1)
	e.RunElevatorProgram("15657", id, numFloors, initTargetFloor)
	fmt.Println(e)
}

func main() {
	testElevator()
}