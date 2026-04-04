package session

import (
	"elevator_program/message"
	"elevator_program/udp/packet"
	"fmt"
)

func (ses *Session) GenerateDataPacket(
	senderAddr string,
	pktType packet.PacketType,
	eMsg message.ElevatorMessage,
) ([]byte, error) {
	pkt := packet.Packet{
		Header: packet.Header{
			Seq:           ses.seq,
			SessionID:     ses.ID,
			PktType:       pktType,
			RecipientAddr: ses.peerAddr.String(),
			SenderAddr:    senderAddr,
		},
		Payload: eMsg,
	}

	data, err := pkt.EncodePacket()
	if err != nil {
		return nil, err
	}

	return data, nil
}

// for the SessionBehavior, does nothing
func (ses *Session) OnSend(pktType packet.PacketType) {}

func (ses *Session) sendLoop(behavior SessionBehavior) {
	defer ses.wg.Done()

	for {
		select {
		case <-ses.stop:
			return

		case outPkt := <-ses.outgoingMsgCh:
			ses.handleOutPkt(outPkt, behavior)
		}
	}
}

func (ses *Session) handleOutPkt(outPkt outgoingMessage, behavior SessionBehavior) {
	err := ses.send(outPkt)
	if err != nil {
		fmt.Printf("Session %d: send error: %v\n", ses.ID, err)
		return
	}

	if outPkt.PktType == packet.PKT_T_IAmMaster {
		ses.setSelfAsMaster()
		ses.setIsSynced(true)
	}

	behavior.OnSend(outPkt.PktType)
}
