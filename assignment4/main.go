package main

import (
	"assignment4/process_pair"
	"fmt"
	"os/exec"
	"strconv"
)

func startNewProcess(newID int, udpAddr string) {
	// Pass the backup ID to the new process
	cmd := exec.Command("./main", "backup", strconv.Itoa(newID), udpAddr)
	err := cmd.Start()
	if err != nil {
		fmt.Println("Failed to start new backup process:", err)
	} else {
		fmt.Println("Started new backup process with ID", newID)
	}
}

func takeover(id int, udpAddr string, startCount int) {
	fmt.Println("Backup taking over as primary")
	// spawn new backup first
	startNewProcess(id+1, udpAddr) // then start new primary
	process_pair.Primary(id, udpAddr, startCount)
}

func main() {
	// udpAddr := "10.100.23.15:30000" // Use a port for testing
	udpAddr := "127.0.0.1:9999" // Loopback
	id := 1

	ready := make(chan struct{})

	go process_pair.Backup(id, udpAddr, ready, func(startCount int) {
		takeover(id, udpAddr, startCount)
	})

	// wait until backup is ready
	<-ready

	startNewProcess(id+1, udpAddr)
	process_pair.Primary(0, udpAddr, 1) // Start initial primary

	done := make(chan struct{})
	<-done
}
