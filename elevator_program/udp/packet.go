package udp

import (
	"encoding/json"
	"fmt"
	"net"
)

type Message struct {
	Content string `json:"content,omitempty"`
}

type incommingPacket struct {
	packet Packet
	addr   *net.UDPAddr
}

type Packet struct {
	Header  Header  `json:"header"`
	Payload Message `json:"payload"`
}

// type incomingPacket struct {
// 	packet Packet
// 	addr   *net.UDPAddr
// }

func encodePacket(pck Packet) []byte {
	data, err := json.Marshal(pck)
	if err != nil {
		panic(err)
	}
	return data
}

func decodePacket(buf []byte, n int) Packet {
	var pck Packet
	err := json.Unmarshal(buf[:n], &pck)
	if err != nil {
		panic(err)
	}
	return pck
}

// helper to encode & send a packet
func sendPacket(conn *net.UDPConn, addr *net.UDPAddr, pck Packet) error {
	data := encodePacket(pck)
	_, err := conn.WriteToUDP(data, addr)
	if err != nil {
		fmt.Println("Send error:", err)
		return err
	}

	return nil
}
