package main

import (
	// "fmt"
	"elevator_program/elevator"
	// "elevator_program/utilities"
	// "elevator_program/elevio"
)

func testElevator() {
	var e elevator.Elevator
	
	id := 1
	numFloors := 4
	initFloor := 0

	// fmt.Println(e)
	// e.InitElevator(1)
	e.RunElevatorProgram("15657", id, numFloors, initFloor)
	select{}
}

func main() {
	testElevator()
}