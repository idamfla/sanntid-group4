package main

import (
	"elevator_program/elevator"
	"elevator_program/udp"
	"elevator_program/udp/message"
	"elevator_program/udp/server"
	"elevator_program/udp/session"
	"elevator_program/config"
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
	cfg := config.ParseFlags()

	if cfg.ID == 0 {
		fmt.Println("STARTING ELEVATOR LAUNCHER")
		config.SpawnElevators(cfg)
		select {}
	}

	fmt.Printf(
		"STARTING ONE ELEVATOR: id=%d addr=%s:%s floors=%d initfloor=%d\n",
		cfg.ID, cfg.IP, cfg.Port, cfg.Floors, cfg.InitFloor,
	)

	config.RunOneElevator(cfg)
}


/*
	ch := make(chan session.ElevatorPacket, 32)
	udpPort := 9000 + (cfg.ID - 1)

	srv := createServer(udpPort, fmt.Sprintf("%d", cfg.ID), cfg.N, ch)
	srv.Start()

	go testBroadcast_send(srv)

	for msg := range ch {
		fmt.Println("msgCh test:", msg.Packet.Payload.Content)
		if msg.Done != nil {
			msg.Done <- struct{}{}
		}
	}
}



func main() {

    cfg := config.ParseFlags()

	if cfg.ID == 0 {
		config.SpawnElevators(cfg)
		return
	}

	config.RunOneElevator(cfg)

	chA := make(chan session.ElevatorPacket)
	serverA := createServer(9000, "A", 5, chA)

	chB := make(chan session.ElevatorPacket)
	serverB := createServer(9001, "B", 2, chB)

	serverA.Start()
	serverB.Start()

	// serverA.StartSession(udp.MustUDPAddr("127.0.0.1", 9001),
	serverA.StartBroadcast(message.Message{Content: "Hello from A"})
	// 	message.Message{Content: "Hello from A"})

	// go testElevator()
	// go testServer(serverA, serverB)

	for msg := range chB {
		fmt.Println("msgCh test:", msg.Packet.Payload.Content)
		if msg.Done != nil {
			msg.Done <- struct{}{}
		}
	}

	closeProgram(serverA, serverB)
}
*/