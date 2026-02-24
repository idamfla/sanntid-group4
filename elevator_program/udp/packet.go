package udp

import (
	"encoding/json"
	"net"
)

type Message struct {
	Content string `json:"content,omitempty"`
}

type Packet struct {
	Header  Header  `json:"header"`
	Payload Message `json:"payload"`
}

type incomingPacket struct {
	packet Packet
	addr   *net.UDPAddr
}

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
