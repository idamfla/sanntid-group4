package process_pair

import (
	"fmt"
	"net"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

const (
	heartbeatInterval = 1 * time.Second
	backupTimeout     = 3 * time.Second
	primaryLifetime   = 5 * time.Second
)

func Backup(myAddr string) {
	fmt.Println("Backup listening on", myAddr)

	conn := mustListenUDP(myAddr)
	defer conn.Close()

	buf := make([]byte, 1024)
	lastCount := 1

	for {
		conn.SetReadDeadline(time.Now().Add(backupTimeout))

		n, _, err := conn.ReadFromUDP(buf)
		if err != nil {
			fmt.Println("Backup: primary dead, taking over")
			takeover(lastCount, myAddr)
			return
		}

		num, err := strconv.Atoi(strings.TrimSpace(string(buf[:n])))
		if err != nil {
			continue
		}

		lastCount = num
		fmt.Println("Backup: heartbeat", lastCount)
	}
}

func Primary(backupAddr string, startCount int) {
	fmt.Println("Primary sending to", backupAddr)

	conn := mustDialUDP(backupAddr)
	defer conn.Close()

	count := startCount
	start := time.Now()

	for {
		if time.Since(start) > primaryLifetime {
			fmt.Println("Primary: simulating crash")
			return
		}

		sendHeartbeat(conn, count)
		fmt.Println("Primary heartbeat", count)

		count = nextCount(count)

		time.Sleep(1 * time.Second)
	}
}

func takeover(startCount int, myAddr string) {
	newBackupAddr := incrementPort(myAddr)

	fmt.Println("Starting new backup on", newBackupAddr)

	cmd := newBackupCommand(newBackupAddr) // <- spawns new backup
	if err := cmd.Start(); err != nil {
		panic(err)
	}

	time.Sleep(200 * time.Millisecond)

	Primary(newBackupAddr, startCount)
}

// region Helpers
func mustListenUDP(addr string) *net.UDPConn {
	udpAddr, err := net.ResolveUDPAddr("udp4", addr)
	if err != nil {
		panic(err)
	}

	conn, err := net.ListenUDP("udp4", udpAddr)
	if err != nil {
		panic(err)
	}
	return conn
}

func mustDialUDP(addr string) *net.UDPConn {
	udpAddr, err := net.ResolveUDPAddr("udp4", addr)
	if err != nil {
		panic(err)
	}

	conn, err := net.DialUDP("udp4", nil, udpAddr)
	if err != nil {
		panic(err)
	}

	return conn
}

func sendHeartbeat(conn *net.UDPConn, count int) {
	msg := strconv.Itoa(count)
	if _, err := conn.Write([]byte(msg)); err != nil {
		fmt.Println("Primary send failed:", err)
	}
}

func nextCount(count int) int {
	if count >= 4 {
		return 1
	}
	return count + 1
}

func incrementPort(addr string) string {
	host, portStr, _ := net.SplitHostPort(addr)
	port, _ := strconv.Atoi(portStr)
	return host + ":" + strconv.Itoa(port+1)
}

func newBackupCommand(addr string) *exec.Cmd {
	_, portStr, _ := net.SplitHostPort(addr)
	cmd := exec.Command("./main", "backup", portStr)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd
}

// endregion
