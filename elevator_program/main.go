package main

import (
	"elevator_program/coordinator"
	"elevator_program/elevator"
	"elevator_program/message"
	elevtest "elevator_program/udp/elev_test"
	"elevator_program/udp/packet"
	"elevator_program/udp/server"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	// "elevator_program/utilities"
	"elevator_program/elevio"
)

const (
	localIP  = "10.100.23.27"
	receiver = "10.100.23.27"
)

func testElevator() {
	var e elevator.Elevator
	// fmt.Println(e)

	id := "A"
	numFloors := 4
	initFloor := 3 // NB! in the code the elevator floors are 0-index, on the controller it is not
	ip_address := "localhost"
	port := "15657"

	// "localhost:15657"
	elevio.Init(ip_address+":"+port, numFloors)

	e.InitElevator(id, numFloors, initFloor, ip_address, 5000) // TODO WHAT TO DO HERE, prot is int??
	e.RunElevatorProgram()

	p := coordinator.Coordinator{}
	p.InitCoordinator()
	// p.StartServer(ip_address, 5000, id)
	p.Start(&e)
	// TODO MAybe the right spot to put it
	defer p.Close()

	// go p.TestMsgHandler(&e, numFloors)
	// go p.TestMsgHandler_Master(&e, numFloors)

	// e.TestMasterLogic()

	/*
		TODO, bug - when cab to floor 2, then cab to floor 1, if floor 3 is pressed after reaching floor 2, elevator will go up to floor 3
	*/
	select {}
}

func testBroadcast_send(srv *server.Server) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	bcMsg := message.ElevatorMessage{Ip: "Hello, broadcast from " + srv.ID}

	for range ticker.C {
		srv.QueueMessage(nil, packet.PROTO_PKT_T_BroadcastUpdate, bcMsg)
		fmt.Println("bcMsg:", srv.ID, ",", bcMsg)
	}
}

// just to get some prints after i "shut down"
func closeProgram(e1 *elevtest.Elev, e2 *elevtest.Elev) {
	// Create signal channel
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	// Wait for Ctrl+C
	<-sigChan

	fmt.Println("\nCtrl+C pressed")

	// Graceful shutdown
	e1.Close()
	e2.Close()

	fmt.Println("Servers shut down cleanly")
}

func main() {
	ip_address := receiver
	port := "15657"

	elevio.Init(ip_address+":"+port, 4)

	e1 := elevator.Elevator{}
	e1.InitElevator("1", 4, 3, receiver, 9005)
	p1 := coordinator.Coordinator{}
	p1.InitCoordinator()
	err := p1.StartServer(receiver, 9005, "1")
	if err != nil {
		fmt.Println(err)
		return
	}
	p1.Start(&e1)

	// fmt.Println("eeeee: ", e1)

	fmt.Println("Created e1 and p1")

	// e2 := elevator.Elevator{}
	// e2.InitElevator("2", 4, 3, localIP, 9001)
	// p2 := coordinator.Coordinator{}
	// p2.InitCoordinator()
	// err = p2.StartServer(localIP, 9001, "2")
	// if err != nil {
	// 	fmt.Println(err)
	// 	return
	// }

	// p2.Start(&e2)

	fmt.Println(&e1)
	// fmt.Println(&e2)
	fmt.Println("Am i stuck")

	e1.RunElevatorProgram()

	// go func() {
	// 	time.Sleep(15 * time.Second)

	// 	fmt.Println("\n=== TEST 3: Elevator failed / motor stop ===")
	// 	e1.FaultMsg <- message.FaultMessage{
	// 		FaultType: message.FAULT_T_ElevatorFailed,
	// 	}

	// }()
	select {}

	//TODO
	/*
		make sure elevator_task sendElevatorTaskLoop works correctly ...
		there is an issue after the quorum is reached in the broadcast_session, it stops there
		fix, the one that broadcasts has no content inside the task ...
	*/
}

// 127.0.0.1 er lokal <- du kan bruke denne for en til en
// 10.22.67.255 broadcast, broadcast har 255 som siste verdi

// TODO test go run -race

// TODO it looks like the system gets multiple running hallrequest and forgets its own target
// TODO ask if there is some logical errors in the scanning of requests, sorting them and delegating them
