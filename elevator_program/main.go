package main

import (
	// "fmt"
	"elevator_program/elevator"
	"elevator_program/udp"
	"elevator_program/udp/message"
	"elevator_program/udp/server"
	"elevator_program/udp/session"
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
	select {}
}

func createServer(port int, id string, numberOfElevators int, ch1 chan<- session.ElevatorPacket) *server.Server {
	s1, err := server.NewServer(localIP, port, id, numberOfElevators, ch1)
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

	aMsg := message.Message{Content: "Hello from A"}
	bMsg := message.Message{Content: "Hello from B"}

	// Server A sends to B
	go s1.StartSession(udp.MustUDPAddr("127.0.0.1", 9001), aMsg)

	// Server B sends to A
	go s2.StartSession(udp.MustUDPAddr("127.0.0.1", 9000), bMsg)
}

func testBroadcast_send(srv *server.Server) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	bcMsg := message.Message{Content: "Hello, broadcast from " + srv.ID}

	for range ticker.C {
		srv.StartBroadcast(bcMsg)
		fmt.Println("bcMsg:", srv.ID, ",", bcMsg)
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
	chA := make(chan session.ElevatorPacket)
	serverA := createServer(9000, "A", 2, chA)

	chB := make(chan session.ElevatorPacket)
	serverB := createServer(9001, "B", 2, chB)

	// serverA.StartSession(udp.MustUDPAddr("127.0.0.1", 9001),
	// 	message.Message{Content: "Hello from A"})
	serverA.StartBroadcast(message.Message{Content: "Hello from A"})

	// go testServer(serverA, serverB)

	serverA.Start()
	serverB.Start()

	// bcMsg := "Hello, broadcast from " + "A"
	// serverA.SendBroadcast(1, 1, message.Message{Content: bcMsg})

	// go testBroadcast_send(serverA)

	for msg := range chB {
		fmt.Println("msgCh test:", msg.Packet.Payload.Content)
		if msg.Done != nil {
			msg.Done <- struct{}{}
		}
	}

	closeProgram(serverA, serverB)
}
