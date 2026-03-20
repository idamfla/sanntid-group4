package udp

import (
	"fmt"
	"net"
	"time"
)

const (
	READ_INTERVAL            = 500 * time.Millisecond
	RETRY_INTERVAL           = 500 * time.Millisecond
	MAX_RETRIES              = 10
	SHUTDOWN_TIMEOUT         = 5 * time.Second
	LOCAL_COMMIT_TIMEOUT     = 5 * time.Second
	BROADCAST_ACK_TIMEOUT    = 2 * time.Second
	BROADCAST_COMMIT_TIMEOUT = 3 * time.Second
	TASK_READY_TIMEOUT       = 2 * time.Second
	MASTER_ELECTION_TIMEOUT  = 500 * time.Millisecond

	OtherBroadcastIP = "172.20.10.15"
	NtnuBroadcastIP  = "10.22.119.255"
	HomeBroadcastIP  = "192.168.50.255"
	BroadcastIP      = NtnuBroadcastIP
	BROADCAST_PORT   = 3000
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
