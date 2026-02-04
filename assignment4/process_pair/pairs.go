package process_pair

import (
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"
)

func Backup(id int, udpAddr string, takeOverCallback func(startCount int)) {
	lastCount := 1

	addr, _ := net.ResolveUDPAddr("udp4", udpAddr)
	conn, _ := net.ListenUDP("udp4", addr)
	defer conn.Close()

	buf := make([]byte, 1024)

	timeout := 3 * time.Second

	for {
		conn.SetReadDeadline(time.Now().Add(timeout))
		n, _, err := conn.ReadFromUDP(buf)
		if err != nil {
			fmt.Printf("Backup %d: Primary dead, taking over\n", id)
			takeOverCallback(lastCount)
			return
		}

		num, _ := strconv.Atoi(strings.TrimSpace(string(buf[:n])))
		lastCount = num
		fmt.Printf("Backup %d: Primary alive, heartbeat %d\n", id, lastCount)
	}
}

func Primary(id int, udpAddr string, startCount int) {
	startTime := time.Now()
	threshold := 5 * time.Second

	count := startCount

	addr, _ := net.ResolveUDPAddr("udp4", udpAddr)
	conn, _ := net.DialUDP("udp4", nil, addr)
	defer conn.Close()

	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		// msg := fmt.Sprintf("Primary: heartbeat %d\n", id)
		if time.Since(startTime) > threshold {
			return
		}

		msg := fmt.Sprintf("%d", count)
		_, err := conn.Write([]byte(msg))
		if err != nil {
			fmt.Println("Primary failed to send heartbeat:", err)
			return
		}

		if count < 4 {
			count += 1
		} else {
			count = 1
		}
		time.Sleep(1 * time.Second)
	}
}
