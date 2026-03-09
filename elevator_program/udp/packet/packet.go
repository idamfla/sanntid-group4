package packet

import (
	"elevator_program/udp/message"
	"encoding/json"
	"fmt"
	"net"
)

type Packet struct {
	Header  Header          `json:"header"`
	Payload message.Message `json:"payload"`
}

func (p Packet) encodePacket() ([]byte, error) {
	return json.Marshal(p)
}

func DecodePacket(buf []byte, n int) (Packet, error) {
	var pck Packet
	err := json.Unmarshal(buf[:n], &pck)
	if err != nil {
		return Packet{}, err
	}
	return pck, nil
}

// helper to encode & send a packet
func SendPacket(conn *net.UDPConn, addr *net.UDPAddr, pck Packet) error {
	data, err := pck.encodePacket()
	if err != nil {
		fmt.Println("Encode error:", err)
		return err
	}

	_, err = conn.WriteToUDP(data, addr)
	if err != nil {
		fmt.Println("Send error:", err)
		return err
	}

	return nil
}
