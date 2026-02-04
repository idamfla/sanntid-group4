package main

import (
	"assignment4/process_pair"
	"fmt"
	"net"
	"os/exec"
	"strconv"
	"time"
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

func IncrementUDPPort(udpAddr string) (string, error) {
	host, portStr, err := net.SplitHostPort(udpAddr)
	if err != nil {
		return "", fmt.Errorf("invalid UDP address: %w", err)
	}

	port, err := strconv.Atoi(portStr)
	if err != nil {
		return "", fmt.Errorf("invalid port: %w", err)
	}

	newPort := port + 1
	newAddr := net.JoinHostPort(host, strconv.Itoa(newPort))
	return newAddr, nil
}

func takeover(id int, udpAddr string, startCount int) {
	fmt.Println("Backup taking over as primary")

	newAddr, _ := IncrementUDPPort(udpAddr)

	// spawn new backup first
	startNewProcess(id+1, newAddr)
	time.Sleep(100 * time.Millisecond)
	go process_pair.Primary(id, newAddr, startCount)
}

func main() {
	// udpAddr := "10.100.23.15:30000" // Use a port for testing
	udpAddr := "127.0.0.1:9999" // Loopback
	id := 1

	go process_pair.Backup(id, udpAddr, func(startCount int) {
		takeover(id, udpAddr, startCount)
	})

	startNewProcess(id+1, udpAddr)
	process_pair.Primary(0, udpAddr, 1) // Start initial primary
	time.Sleep(100 * time.Millisecond)
	done := make(chan struct{})
	<-done
}
