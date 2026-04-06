package session

import (
	"elevator_program/message"
	"elevator_program/udp"
	"elevator_program/udp/packet"
	"fmt"
)

type WhoIsAliveBroadcast struct {
	*BaseBroadcastSession
	election *Election
}

func NewWhoIsAliveBroadcast(id uint32, srv ServerAPI) *WhoIsAliveBroadcast {
	ws := &WhoIsAliveBroadcast{
		BaseBroadcastSession: NewBaseBroadcastSession(id, srv, 0),
		election:             &Election{masterFound: make(chan struct{}, 1)},
	}
	return ws
}

func (ws *WhoIsAliveBroadcast) Start() {
	ws.wg.Add(2)
	go ws.listen(ws)
	go ws.sendLoop(ws)

	ws.queueWhoIsAliveMsg()
}

func (ws *WhoIsAliveBroadcast) Close() {
	ws.BaseBroadcastSession.Close()
}

func (ws *WhoIsAliveBroadcast) GetID() uint32 { return ws.BaseBroadcastSession.GetID() }

func (ws *WhoIsAliveBroadcast) ReceivePacket(pkt packet.Packet) {
	ws.BaseBroadcastSession.ReceivePacket(pkt)
}

func (ws *WhoIsAliveBroadcast) QueueDirectMsg(pktType packet.PacketType, outMsg packet.OutgoingMessage) { // TODO this should not exsist outside of session ...
	ws.BaseBroadcastSession.QueueDirectMsg(pktType, outMsg)
}

func (ws *WhoIsAliveBroadcast) OnSend(pktType packet.PacketType) {
	switch pktType {
	case packet.PKT_T_WhoIsAlive:
		ws.startElection()
	case packet.PKT_T_IAmMaster:
		ws.startResponseTimer()
	}
}

func (ws *WhoIsAliveBroadcast) HandleIncPkt(pkt packet.Packet) error {
	peerKey := pkt.Header.SenderAddr
	pktType := pkt.Header.PktType
	ws.addResponder(peerKey)

	switch pktType {
	case packet.PKT_T_WhoIsAlive:
		ws.handleWhoIsAlive()

	case packet.PKT_T_IAmAlive:
		fmt.Println("from:", peerKey, pktType)

	case packet.PKT_T_IAmMaster:
		ws.handleIAmMaster()

	case packet.PKT_T_ElectedMasterIs:
		ws.handleElectedMasterIs(pkt)

	case packet.PKT_T_MasterAck:
		ws.handleMasterAck()
	}

	return nil
}

func (ws *WhoIsAliveBroadcast) handleWhoIsAlive() {
	if ws.isMaster() {
		ws.queueIamMasterMsg()
	} else {
		ws.queueReply(packet.PKT_T_IAmAlive)
	}
}

func (ws *WhoIsAliveBroadcast) handleIAmMaster() {
	select {
	case ws.election.masterFound <- struct{}{}:
	default:
	}

	ws.queueElevatorCommand(message.EMSG_T_IAmMaster) // TODO how should task look
	ws.clearHasLastPacket()
	ws.queueReply(packet.PKT_T_MasterAck)
}

func (ws *WhoIsAliveBroadcast) handleElectedMasterIs(pkt packet.Packet) {
	addr := pkt.Payload.Addr
	if addr == ws.selfAddr {
		fmt.Println("I was elected master", ws.selfAddr) // TODO db, remove
		ws.queueIamMasterMsg()
	}
}

func (ws *WhoIsAliveBroadcast) handleMasterAck() {
	if ws.isMaster() {
		fmt.Printf("MstrAck: %d/%d\n", ws.countResponders(), ws.expectedResponses)

		if ws.countResponders() >= ws.expectedResponses {
			ws.clearHasLastPacket()
			ws.stopResponseTimer()
		}
	}
}

func (ws *WhoIsAliveBroadcast) startResponseTimer() {
	ws.responseTimer.Restart(udp.RESPONSE_TIMEOUT, func() {
		if ws.countResponders() >= ws.expectedResponses {
			// Ignore false timeout
			return
		}

		fmt.Printf("Peer(s) did not respond to masterElectSession in time ... %d/%d\n", ws.countResponders(), ws.expectedResponses)
		ws.queueWhoIsAliveMsg()
		ws.stopResponseTimer()
	})
}

func (ws *WhoIsAliveBroadcast) queueElectedMasterMsg(masterAddr string) { // TODO queue at server
	ws.srv.QueueElectedMasterMsg(masterAddr)
}

func (ws *WhoIsAliveBroadcast) startElection() { ws.election.Start(ws) }

func (ws *WhoIsAliveBroadcast) runElection() (string, error) {
	ws.mu.Lock()
	defer ws.mu.Unlock()

	lowest := ""
	for addr := range ws.responders {
		select {
		case <-ws.stop:
			err := fmt.Errorf("Election abortet since session stopped")
			return lowest, err
		default:
		}

		if lowest == "" || addr < lowest {
			lowest = addr
		}
	}
	return lowest, nil
}
