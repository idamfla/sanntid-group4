package main

import (
	"elevator_program/elevator"
	"elevator_program/message"
	"elevator_program/protocol"
	"elevator_program/udp"
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
	localIP  = "127.0.0.1"
	receiver = "10.100.23.15"
)

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

	e.InitElevator(id, numFloors, initFloor, ip_address, port)
	e.RunElevatorProgram()

	p := protocol.Protocol{}
	numElevators := 3                                   // TODO how can I know this before talking to the others?
	p.InitProtocol(ip_address, 5000, "1", numElevators) // TODO fix the values here
	// TODO MAybe the right spot to put it
	defer p.Server.Close()

	go p.MessageListener(&e)
	go p.TestMsgHandler(&e, numFloors)
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

	bcMsg := message.Message{Ip: "Hello, broadcast from " + srv.ID}

	for range ticker.C {
		srv.QueueMessage(nil, packet.PROTO_PKT_T_BroadcastData, bcMsg)
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
	eA := elevtest.NewElev("A", 2)

	err := eA.StartServer(localIP, 9000)
	if err != nil {
		fmt.Println(err)
		return
	}

	eB := elevtest.NewElev("B", 2)

	err = eB.StartServer(localIP, 9001)
	if err != nil {
		fmt.Println(err)
		return
	}

	eA.Start()
	eB.Start()

	eA.QueueMessage(
		udp.MustUDPAddr(localIP, 9001),
		packet.PROTO_PKT_T_BroadcastData,
		message.Message{Ip: "Hello A!"},
	)
	closeProgram(eA, eB)

	//TODO make sure the server dosent take in peers and not totalnumberofElevators
}
