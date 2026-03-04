package main

import (
	// "fmt"
	"elevator_program/elevator"
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
	myIP     = "192.168.50.97"
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
	/*
		TODO, bug - when cab to floor 2, then cab to floor 1, if floor 3 is pressed after reaching floor 2, elevator will go up to floor 3
	*/
	select {}
}

func createServer(port int, id string) *server.Server {
	s1, err := server.NewServer(localIP, port, id)
	if err != nil {
		panic(fmt.Sprintf("Failed to create s1: %v", err))
	}
	if s1 == nil {
		panic("s1 is nil")
	}

	// s2, err := server.NewServer(myIP, 9001, "B")
	// if err != nil {
	// 	panic(fmt.Sprintf("Failed to create s2: %v", err))
	// }
	// if s2 == nil {
	// 	panic("s2 is nil")
	// }

	fmt.Println("Server", id, "is running...")

	// Give them a moment to start
	time.Sleep(time.Second)
	return s1
}

func testServer(s1 *server.Server, s2 *server.Server) {
	// var GlobalServers = make(map[string]*server.Server) // key = "name" or "ip:port"

	// Server A sends to B
	go s1.SendSession(1, "127.0.0.1", 9001, "Hello from A")

	// Server B sends to A
	go s2.SendSession(2, "127.0.0.1", 9000, "Hello from B")
}

func testBroadcast_send(srv *server.Server) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	seq := uint32(0)
	sessionID := uint32(1) // or whatever you need

	bcMsg := "Hello, broadcast from " + srv.ID

	for range ticker.C {
		err := srv.SendBroadcast(seq, sessionID, bcMsg)
		if err != nil {
			fmt.Println("Broadcast error:", err)
		} else {
			fmt.Println("bcMsg:", srv.ID, ",", bcMsg)
		}
		seq++ // increment sequence if needed
		sessionID++
	}
}

func closeProgram(s1 *server.Server, s2 *server.Server) {
	// Create signal channel
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	// Wait for Ctrl+C
	<-sigChan

	fmt.Println("\nCtrl+C pressed")

	// Print session counts
	s1.PrintSessions()
	s2.PrintSessions()

	// Graceful shutdown
	s1.Close()
	s2.Close()

	fmt.Println("Servers shut down cleanly")
}

func main() {
	serverA := createServer(9000, "A")
	serverB := createServer(9001, "B")

	// go serverA.SendSession(1, "127.0.0.1", 9001, "Hello from A")

	// go testServer(serverA, serverB)

	go serverA.Listen()
	go serverB.Listen()

	// bcMsg := "Hello, broadcast from " + "A"
	// serverA.SendBroadcast(1, 1, bcMsg)

	go testBroadcast_send(serverA)

	closeProgram(serverA, serverB)
}
