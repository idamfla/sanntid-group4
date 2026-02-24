package udp

import (
	"fmt"
	"net"
	"time"
)

func SendSession(sessionID uint32, message string) {
	serverAddr, _ := net.ResolveUDPAddr("udp", "127.0.0.1:9000")
	conn, err := net.DialUDP("udp", nil, serverAddr)
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	seq := uint32(1)

	// Send Data
	dataPacket := Packet{
		Header: Header{
			Seq:       seq,
			MsgType:   MSG_T_Data,
			SessionID: sessionID,
		},
		Payload: Message{Content: message},
	}

	sendPacket(conn, dataPacket)
	fmt.Printf("Sent to session %d: %s\n", sessionID, message)

	// Wait a moment for ACK (optional)
	time.Sleep(time.Second)

	// Send Done
	donePacket := Packet{
		Header: Header{
			Seq:       seq + 1,
			MsgType:   MSG_T_Done,
			SessionID: sessionID,
		},
		Payload: Message{Content: "Done"},
	}

	sendPacket(conn, donePacket)
	fmt.Printf("Sent DONE to session %d\n", sessionID)
}

// helper to encode & send a packet
func sendPacket(conn *net.UDPConn, pck Packet) {
	data := encodePacket(pck)
	_, err := conn.Write(data)
	if err != nil {
		fmt.Println("Send error:", err)
	}
}
