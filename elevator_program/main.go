package main

import (
	"elevator_program/coordinator"
	"elevator_program/elevator"
	"elevator_program/elevio"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"
	"strconv"
	"syscall"
)

func closeProgram(c *coordinator.Coordinator) {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	<-sigChan

	fmt.Println("\nCtrl+C pressed")

	c.Close()

	fmt.Println("Servers shut down cleanly")
}

func main() {
	id := flag.String("id", "A", "Elevator ID")
	simPort := flag.Int("simport", 15657, "Simulator TCP port")
	udpPort := flag.Int("port", 9000, "UDP port for this elevator")
	localIP := flag.String("ip", "127.0.0.1", "Local IP address")
	flag.Parse()

	// Log to both terminal and file logs/elevator_<id>.log
	os.MkdirAll("logs", 0755)
	logFile, err := os.Create(fmt.Sprintf("logs/elevator_%s.log", *id))
	if err != nil {
		log.Fatal("Could not create log file:", err)
	}
	defer logFile.Close()

	// Create a pipe to capture all fmt.Print* output
	origStdout := os.Stdout
	pr, pw, _ := os.Pipe()
	os.Stdout = pw

	// Tee pipe output to both original terminal and log file
	go func() {
		mw := io.MultiWriter(origStdout, logFile)
		io.Copy(mw, pr)
	}()

	portstr := strconv.Itoa(*udpPort)

	elevio.Init(fmt.Sprintf("localhost:%d", *simPort), 4)

	e1 := elevator.Elevator{}
	e1.InitElevator(*id, 4, 3, *localIP+":"+portstr, *udpPort)
	c1 := coordinator.Coordinator{}
	c1.InitCoordinator()
	err = c1.StartServer(*localIP, *udpPort, *id)
	if err != nil {
		fmt.Println(err)
		return
	}
	c1.Start(&e1)

	fmt.Println(&e1)

	defer closeProgram(&c1)

	e1.RunElevatorProgram()

	// msg := message.ElevatorMessage{
	// 	EMsgType: message.EMSG_T_NewToChannel,
	// 	ID:       e1.Id,
	// }
	// e1.SendToCoordinator <- msg
}
