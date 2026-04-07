package udp

import (
	"fmt"
	"net"
	"time"
)

const (
	RETRY_INTERVAL          = 500 * time.Millisecond
	MAX_RETRIES             = 10
	SHUTDOWN_TIMEOUT        = 5 * time.Second
	LOCAL_COMMIT_TIMEOUT    = 5 * time.Second
	RESPONSE_TIMEOUT        = 2 * time.Second
	TASK_READY_TIMEOUT      = 2 * time.Second
	MASTER_ELECTION_TIMEOUT = 500 * time.Millisecond
	// PEER_TIMEOUT             = 15

	// --- IP and Port ---
	// Group4IP        = "10.100.23.15"
	// NtnuBroadcastIP = "10.22.67.255"
	// HomeBroadcastIP = "192.168.86.255"
	// BroadcastIP     = HomeBroadcastIP
	OtherBroadcastIP    = "172.20.10.15"
	NtnuBroadcastIP     = "10.100.23.255"
	HomeBroadcastIP     = "192.168.10.255"
	LoopbackBroadcastIP = "127.255.255.255"
	BroadcastIP         = NtnuBroadcastIP // Change to HomeBroadcastIP for network
	BROADCAST_PORT      = 3000
)

func IPPortToUDPAddr(ip string, port int) (*net.UDPAddr, error) {
	addr := fmt.Sprintf("%s:%d", ip, port)
	return net.ResolveUDPAddr("udp", addr)
}

func StringAddrToUDPAddr(addr string) (*net.UDPAddr, error) {
	return net.ResolveUDPAddr("udp", addr)
}

func MustUDPAddr(ip string, port int) *net.UDPAddr {
	udpAddr, err := IPPortToUDPAddr(ip, port)
	if err != nil {
		panic(err)
	}
	return udpAddr
}
