package main

import (
	// "fmt"
	"elevator_program/elevator"
	elevtest "elevator_program/udp/elev_test"
	"elevator_program/udp/message"
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

	e.InitElevator(id, numFloors, initFloor)
	e.RunElevatorProgram()
	select {}
}

func testBroadcast_send(srv *server.Server) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	bcMsg := message.Message{Content: "Hello, broadcast from " + srv.ID}

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

	fmt.Println("\nCtrl+C pressed")
	fmt.Println("Is A master", e1.IsMaster())

	// Graceful shutdown
	e1.Close()
	e2.Close()
	e3.Close()

	fmt.Println("Servers shut down cleanly")
}

func main() {
	eA := elevtest.NewElev("A")

	err := eA.StartServer(localIP, 9000) // TODO something here dosent work anymore
	if err != nil {
		fmt.Println(err)
		return
	}

	eB := elevtest.NewElev("B")

	err = eB.StartServer(localIP, 9001)
	if err != nil {
		fmt.Println(err)
		return
	}

	eC := elevtest.NewElev("C")

	err = eC.StartServer(localIP, 9002)
	if err != nil {
		fmt.Println(err)
		return
	}

	eA.Start()
	eB.Start()
	eC.Start()

	// eB.QueueMessage(
	// 	nil,
	// 	packet.PROTO_PKT_T_WhoIsMaster,
	// 	message.Message{},
	// )
	// eC.QueueMessage(
	// 	udp.MustUDPAddr(localIP, 9001),
	// 	packet.PROTO_PKT_T_SlaveUpdate,
	// 	message.Message{Content: "Hello from A!"},
	// )

	closeProgram(eA, eB, eC)
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
*/
