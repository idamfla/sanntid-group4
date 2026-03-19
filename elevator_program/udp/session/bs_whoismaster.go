package session

import (
	"elevator_program/message"
	"elevator_program/udp/packet"
	"fmt"
	"net"
)

type WhoIsMasterBroadcast struct {
	*BaseBroadcastSession
	election *Election
	// electionStarted bool
	// masterFound     chan struct{}
}

func NewWhoIsMasterBroadcast(id uint32, selfAddr string, addr *net.UDPAddr, closeReq chan<- uint32, tx PacketSender) *WhoIsMasterBroadcast {
	ws := &WhoIsMasterBroadcast{
		BaseBroadcastSession: NewBaseBroadcastSession(id, selfAddr, addr, closeReq, tx, 0),
		election:             &Election{masterFound: make(chan struct{}, 1)},
		// electionStarted: false,
		// masterFound:     make(chan struct{}, 1),
	}
	return ws
}

func (ws *WhoIsMasterBroadcast) Start() {
	ws.wg.Add(2)
	go ws.listen(ws)
	go ws.sendLoop(ws)
	fmt.Printf("'Who Is Master'-broadcast session %d started\n", ws.ID)
}

func (ws *WhoIsMasterBroadcast) Close() {
	// close(ws.masterFound)
	ws.BaseBroadcastSession.Close()
}

func (ws *WhoIsMasterBroadcast) SendReply(pkt packet.PacketType) { ws.Session.SendReply(pkt) }

func (ws *WhoIsMasterBroadcast) ReceivePacket(pkt packet.Packet) { ws.Session.ReceivePacket(pkt) }

func (ws *WhoIsMasterBroadcast) QueueBroadcastUpdateMsg(eMsg message.ElevatorMessage) {}

func (ws *WhoIsMasterBroadcast) QueueWhoIsMasterMsg() {
	ws.outgoingMsgCh <- outgoingMessage{
		PktType: packet.PKT_T_WhoIsMaster,
		EMsg:    message.ElevatorMessage{},
	}
}

func (ws *WhoIsMasterBroadcast) OnSend(pktType packet.PacketType) {
	switch pktType {
	case packet.PKT_T_WhoIsMaster:
		ws.SendReply(packet.PKT_T_IAmAlive)
	case packet.PKT_T_IAmMaster:
		ws.startAckTimer()
	}
}

func (ws *WhoIsMasterBroadcast) HandlePacket(pkt packet.Packet) error {
	peerID := pkt.Header.SenderAddr
	ws.addResponder(peerID)

	switch pkt.Header.PktType {
	case packet.PKT_T_WhoIsMaster:
		ws.handleWhoIsMaster()

	case packet.PKT_T_IAmAlive:
		fmt.Println(peerID, pkt.Header.PktType)

	case packet.PKT_T_IAmMaster:
		ws.handleIAmMaster()

	case packet.PKT_T_MasterAck:
		if ws.isMaster() {
			ws.handleMasterAck()
		}
	}

	return nil
}

func (ws *WhoIsMasterBroadcast) handleWhoIsMaster() {
	ws.SendReply(packet.PKT_T_IAmAlive)
	ws.election.Start(ws)
	// ws.mu.Lock()
	// if !ws.electionStarted {
	// 	ws.electionStarted = true
	// 	ws.wg.Add(1)
	// 	go ws.electMaster()
	// }

	// ws.mu.Unlock()
}

func (ws *WhoIsMasterBroadcast) handleIAmMaster() {
	select {
	case ws.election.masterFound <- struct{}{}:
	default:
	}
	// select {
	// case ws.masterFound <- struct{}{}:
	// default:
	// }

	ws.SendReply(packet.PKT_T_MasterAck)
	ws.scheduleSessionClose()
}

func (ws *WhoIsMasterBroadcast) handleMasterAck() {
	fmt.Printf("MstrAck: %d/%d\n", ws.countResponders(), ws.expectedResponses)
	ws.startAckTimer()

	if ws.countResponders() >= ws.expectedResponses {
		ws.hasLastPkt = false
		ws.stopAckTimer()
		ws.requestClose()
	}
}

func (ws *WhoIsMasterBroadcast) isMaster() bool {
	return ws.tx.IsMaster()
}

// func (ws *WhoIsMasterBroadcast) electMaster() {
// 	defer ws.wg.Done()

// 	timer := time.NewTimer(udp.MASTER_ELECTION_TIMEOUT)
// 	defer timer.stop()

// 	select {
// 	case <-ws.masterFound:
// 		fmt.Println("Master already exists, stopping election")
// 		return

// 	case <-timer.C:
// 		fmt.Println("No master found, electing...")

// 		ws.mu.Lock()

// 		if len(ws.responders) == 0 {
// 			ws.mu.Unlock()
// 			return
// 		}

// 		lowest := ""
// 		for addr := range ws.responders {
// 			if lowest == "" || addr < lowest {
// 				lowest = addr
// 			}
// 		}

// 		isMaster := lowest == ws.selfAddr

// 		ws.mu.Unlock()

// 		fmt.Println("Lowest:", lowest)

// 		if isMaster {
// 			fmt.Println(ws.selfAddr, "is the new master")
// 			ws.SendReply(packet.PKT_T_IAmMaster)
// 			ws.expectedResponses = ws.countResponders()
// 			ws.resetResponders()
// 		}

// 	case <-ws.stop:
// 		return
// 	}
// }
