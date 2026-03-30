package main

import (
	"elevator_program/coordinator"
	"elevator_program/elevator"
	"elevator_program/elevio"
	"elevator_program/message"
	"fmt"
	"os"
	"os/signal"
	"syscall"
)

const (
	// localIP  = "10.100.23.27"
	localIP  = "127.0.0.1" // "172.20.10.9"
	receiver = "10.100.23.22"
)

func closeProgram(e1 *coordinator.Coordinator, e2 *coordinator.Coordinator) {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	<-sigChan

	fmt.Println("\nCtrl+C pressed")

	e1.Close()
	e2.Close()

	fmt.Println("Servers shut down cleanly")
}

func main() {
	ip_address := "localhost"
	port := "15657"

	elevio.Init(ip_address+":"+port, 4)

	e1 := elevator.Elevator{}
	e1.InitElevator("A", 4, 3, localIP, 9000)
	c1 := coordinator.Coordinator{}
	c1.InitCoordinator()
	err := c1.StartServer(localIP, 9000, "A")
	if err != nil {
		fmt.Println(err)
		return
	}
	c1.Start(&e1)

	fmt.Println(&e1)

	e2 := elevator.Elevator{}
	e2.InitElevator("B", 4, 3, localIP, 9001)
	c2 := coordinator.Coordinator{}
	c2.InitCoordinator()
	err = c2.StartServer(localIP, 9001, "B")
	if err != nil {
		fmt.Println(err)
		return
	}
	c2.Start(&e2)

	fmt.Println(&e2)

	defer closeProgram(&c1, &c2)

	e1.RunElevatorProgram()

	msg := message.ElevatorMessage{
		EMsgType: message.EMSG_T_NewToChannel,
		ID:       e1.Id,
	}
	e1.SendToCoordinator <- msg
}
