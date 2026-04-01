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
}

func NewWhoIsMasterBroadcast(id uint32, selfAddr string, addr *net.UDPAddr, closeReq chan<- uint32, tx ServerAPI) *WhoIsMasterBroadcast {
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
	fmt.Printf("'WIM'-broadcast session %d started\n", ws.ID)
}

func (ws *WhoIsMasterBroadcast) Close() {
	ws.BaseBroadcastSession.Close()
}

func (ws *WhoIsMasterBroadcast) SendReply(pkt packet.PacketType) { ws.Session.SendReply(pkt) }

func (ws *WhoIsMasterBroadcast) ReceivePacket(pkt packet.Packet) { ws.Session.ReceivePacket(pkt) }

func (ws *WhoIsMasterBroadcast) QueueStateBSUpdateMsg(pktType packet.PacketType, eMsg message.ElevatorMessage) {
}

func (ws *WhoIsMasterBroadcast) QueueWhoIsMasterMsg() {
	ws.Session.QueueWhoIsMasterMsg()
}

func (ws *WhoIsMasterBroadcast) OnSend(pktType packet.PacketType) {
	switch pktType {
	case packet.PKT_T_WhoIsMaster:
		ws.election.Start(ws)
		ws.SendReply(packet.PKT_T_IAmAlive)
	case packet.PKT_T_IAmMaster:
		ws.startResponseTimer()
	}
}

func (ws *WhoIsMasterBroadcast) HandlePacket(pkt packet.Packet) error {
	peerID := pkt.Header.SenderAddr
	ws.addResponder(peerID)

	switch pkt.Header.PktType {
	case packet.PKT_T_WhoIsMaster:
		ws.handleWhoIsMaster()

	case packet.PKT_T_IAmAlive:
		fmt.Println("from:", peerID, pkt.Header.PktType)

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
}

func (ws *WhoIsMasterBroadcast) handleIAmMaster() {
	select {
	case ws.election.masterFound <- struct{}{}:
	default:
	}

	ws.SendReply(packet.PKT_T_MasterAck)
	ws.scheduleSessionClose()
}

func (ws *WhoIsMasterBroadcast) handleMasterAck() {
	fmt.Printf("MstrAck: %d/%d\n", ws.countResponders(), ws.expectedResponses)

	if ws.countResponders() >= ws.expectedResponses {
		ws.hasLastPkt = false
		ws.stopResponseTimer()
		ws.requestClose()
	}
}

func (ws *WhoIsMasterBroadcast) countResponders() int {
	return ws.BaseBroadcastSession.countResponders() - 1
}
