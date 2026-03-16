package peer_info

import (
	"net"
	"time"
)

type PeerInfo struct {
	Addr       *net.UDPAddr
	Seq        uint32
	Active     bool
	LastSeen   time.Time
	SessionIDs []uint32
}
