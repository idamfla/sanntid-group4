package main

import (
	// "fmt"
	"elevator_program/elevator"
	"elevator_program/udp"
	"fmt"
	"time"

	// "elevator_program/utilities"
	"elevator_program/elevio"
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
	/*
		TODO, bug - when cab to floor 2, then cab to floor 1, if floor 3 is pressed after reaching floor 2, elevator will go up to floor 3
	*/
	select {}
}

func testServer() {
	var GlobalServers = make(map[string]*udp.Server) // key = "name" or "ip:port"

	serverA, err := udp.NewServer("127.0.0.1", 9000)
	if err != nil {
		panic(err)
	}

	serverB, err := udp.NewServer("127.0.0.1", 9001)
	if err != nil {
		panic(err)
	}

	GlobalServers["A"] = serverA
	GlobalServers["B"] = serverB

	fmt.Println("Server(s) running...")
	// Start listening in goroutines
	go serverA.Listen()
	go serverB.Listen()

	// Give them a moment to start
	time.Sleep(time.Second)

	// Server A sends to B
	go serverA.SendSession(1, "127.0.0.1", 9001, "Hello from A")

	// Server B sends to A
	go serverB.SendSession(2, "127.0.0.1", 9000, "Hello from B")

	// test SendReply
	// remoteAddr, _ := net.ResolveUDPAddr("udp", "127.0.0.1:9001")

	// // create a fake packet (DATA)
	// pck := udp.Packet{
	// 	Header: udp.Header{
	// 		Seq:       1,
	// 		MsgType:   udp.MSG_T_Data,
	// 		SessionID: 42,
	// 	},
	// 	Payload: udp.Message{Content: "Test"},
	// }

	// // send ACK manually
	// serverA.SendReply(remoteAddr, pck, udp.MSG_T_Ack)
}

func main() {
	testServer()

	done := make(chan struct{})
	<-done
}
