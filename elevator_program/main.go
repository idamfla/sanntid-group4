package main

import (
	// "fmt"
	"elevator_program/elevator"
	// "elevator_program/utilities"
	"elevator_program/elevio"
)

func testElevator() {
	var e elevator.Elevator
	// fmt.Println(e)
	
	id := 1
	numFloors := 4
	initFloor := 0
	port := "15657"

	// "localhost:15657"
	elevio.Init("localhost:"+port, numFloors)

	e.InitElevator(id, numFloors, initFloor)
	e.RunElevatorProgram()
	select{}
}

func main() {
	testElevator()
}