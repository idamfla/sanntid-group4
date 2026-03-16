package peer

import (
	"net"
	"time"
)

type NetworkPeer struct {
	Addr       *net.UDPAddr
	Seq        uint32
	Active     bool
	LastSeen   time.Time
	SessionIDs []uint32
}
