package udp

import (
	"fmt"
	"net"
)

const (
	RETRY_FREQUENCY          = 50
	MAX_RETRIES              = 5
	SHUTDOWN_TIMEOUT         = 5
	LOCAL_COMMIT_TIMEOUT     = 5
	REMOTE_COMMIT_TIMEOUT    = 10
	BROADCAST_ACK_TIMEOUT    = 15
	BROADCAST_COMMIT_TIMEOUT = 15

	// Group4IP        = "10.100.23.15"
	NtnuBroadcastIP = "10.22.119.255"
	HomeBroadcastIP = "192.168.50.255"
	BroadcastIP     = NtnuBroadcastIP
	BROADCAST_PORT  = 3000
)

func IPPortToUDPAddr(ip string, port int) (*net.UDPAddr, error) {
	addr := fmt.Sprintf("%s:%d", ip, port)
	return net.ResolveUDPAddr("udp", addr)
}

func MustUDPAddr(ip string, port int) *net.UDPAddr {
	udpAddr, err := IPPortToUDPAddr(ip, port)
	if err != nil {
		panic(err)
	}
	return udpAddr
}
