package main

import (
	"elevator_program/elevator"
	// "elevator_program/utilities"
	// "elevator_program/elevio"
	"fmt"
)

func testElevator() {
	var e elevator.Elevator
	
	fmt.Println(e)
	// e.InitElevator(1)
	e.RunElevatorProgram("15657", 1)
	fmt.Println(e)
}

func main() {
	testElevator()
}