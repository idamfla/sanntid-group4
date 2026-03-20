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
	localIP  = "127.0.0.1"
	receiver = "10.100.23.15"
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

	bcMsg := message.ElevatorMessage{ActivePeers: 80085}

	for range ticker.C {
		srv.QueueMessage(nil, packet.PROTO_PKT_T_BroadcastUpdate, bcMsg)
		fmt.Println("bcMsg:", srv.ID, ",", bcMsg)
	}
}

// just to get some prints after i "shut down"
func closeProgram(e1 *elevtest.Elev, e2 *elevtest.Elev, e3 *elevtest.Elev) {
	// Create signal channel
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	// Wait for Ctrl+C
	<-sigChan
	// pprof.Lookup("goroutine").WriteTo(os.Stdout, 1)

	fmt.Println("\nCtrl+C pressed")
	fmt.Println("Is A master", e1.IsMaster())

	// Graceful shutdown
	e1.Close()
	e2.Close()
	e3.Close()

	fmt.Println("Servers shut down cleanly")
}

func main() {
	ip_address := "localhost"
	port := "15657"

	elevio.Init(ip_address+":"+port, 4)

	e1 := elevator.Elevator{}
	e1.InitElevator("1", 4, 3, localIP, 9000)
	p1 := coordinator.Coordinator{}
	p1.InitCoordinator()
	err := p1.StartServer(localIP, 9000, "1")
	if err != nil {
		fmt.Println(err)
		return
	}
	p1.Start(&e1)

	fmt.Println(&e1)

	e1.RunElevatorProgram()

	// time.Sleep(2 * time.Second)
	msg := message.ElevatorMessage{
		EMsgType: message.EMSG_T_NewToChannel,
		ID:       e1.Id,
	}
	e1.SendToCoordinator <- msg

	select {}
}

// TODO
/*
- make peer sync
- send snapshot msg
- make peer track seq number
- make sure all is orginized enough, folder-vise etc.
- handle new msg, nb some only send forth ack and then do work before starting completely new session
	- check if it is only stopping because it uses old logic, it does send the whole exchange before crashing
- remove master
- should "I am master" prompt you to remove old master if there are any?
- start server with start seq
*/
