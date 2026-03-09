package main

import (
	"elevator_program/elevator"
	"elevator_program/udp"
	"elevator_program/udp/server"
	"fmt"

	// "elevator_program/utilities"
	"elevator_program/elevio"
)

const (
	localhost = "127.0.0.1"
	myIP      = "192.168.50.123"
	receiver  = "10.100.23.15"
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

	// e.TestMasterLogic()

	/*
		TODO, bug - when cab to floor 2, then cab to floor 1, if floor 3 is pressed after reaching floor 2, elevator will go up to floor 3
	*/
	// select {}
}

func main() {
	// cfg := config.ParseFlags()

	// // Launcher mode (no ID given)
	// if cfg.ID == 0 {
	// 	config.SpawnElevators(cfg)
	// 	select {}
	// }

	// // Single elevator mode
	// config.RunOneElevator(cfg)

	go testElevator()

	msgCh := make(chan udp.ElevatorMessage)

	server, err := server.NewServer("1.127.0.0", 9000, "A", msgCh)
	if err != nil {
		fmt.Println("Failed to make server:", err)
	}

	go server.Listen()

	select {}
}
