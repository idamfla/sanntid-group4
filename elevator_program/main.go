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
	// eA := elevtest.NewElev("A", 2)

	// err := eA.StartServer(localIP, 9000) // TODO something here dosent work anymore
	// if err != nil {
	// 	fmt.Println(err)
	// 	return
	// }

	// eB := elevtest.NewElev("B", 2)

	// err = eB.StartServer(localIP, 9001)
	// if err != nil {
	// 	fmt.Println(err)
	// 	return
	// }

	// eA.Start()
	// eB.Start()

	// eA.QueueMessage(
	// 	udp.MustUDPAddr(localIP, 9001),
	// 	packet.PROTO_PKT_T_BroadcastUpdate,
	// 	message.ElevatorMessage{Content: "Hello A!"},
	// )

	// closeProgram(eA, eB)

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

	// fmt.Println("eeeee: ", e1)

	fmt.Println("Created e1 and p1")

	e2 := elevator.Elevator{}
	e2.InitElevator("2", 4, 3, localIP, 9001)
	p2 := coordinator.Coordinator{}
	p2.InitCoordinator()
	err = p2.StartServer(localIP, 9001, "2")
	if err != nil {
		fmt.Println(err)
		return
	}

	p2.Start(&e2)

	fmt.Println(&e1)
	fmt.Println(&e2)

	e2.RunElevatorProgram()

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
