package session

// import (
// 	"elevator_program/udp/packet"
// 	"sync"
// )

// type WhoIsMasterSession struct {
// 	*Session
// 	responders  map[string]bool
// 	masterFound chan struct{}
// 	mu          sync.Mutex
// }

// func (ws *WhoIsMasterSession) HandlePacket(pkt packet.Packet) error {
// 	switch pkt.Header.PktType {
// 	case packet.PKT_T_WhoIsMaster, packet.PKT_T_IAmAlive:
// 		ws.addResponder(pkt.Header.SenderAddr)
// 		// ignore seq
// 	case packet.PKT_T_IAmMaster:
// 		select {
// 		case ws.masterFound <- struct{}{}:
// 		default:
// 		}
// 	}
// 	return nil
// }
