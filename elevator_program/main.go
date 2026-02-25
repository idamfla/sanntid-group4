package main

import (
	"elevator_program/config"
	"elevator_program/elevator"
	"elevator_program/elevio"
)

// Remove this later if we see that the communication works
func testElevator() {
	var e elevator.Elevator
	// fmt.Println(e)

	id := 1
	numFloors := 4
	initFloor := 3 // NB! in the code the elevator floors are 0-index, on the controller it is not
	ip_address := "localhost"
	port := "15657"

	// "localhost:15657"
	elevio.Init(ip_address+":"+port, numFloors)

	e.InitElevator(id, numFloors, initFloor)
	e.RunElevatorProgram()
	/*
		TODO, bug - when cab to floor 2, then cab to floor 1, if floor 3 is pressed after reaching floor 2, elevator will go up to floor 3
	*/
	select {}
}

func main() {
	cfg := config.ParseFlags()

	// Launcher mode (no ID given)
	if cfg.ID == 0 {
		config.SpawnElevators(cfg)
		select {}
	}

	// Single elevator mode
	config.RunOneElevator(cfg)
	select {}
}

//testElevator()
