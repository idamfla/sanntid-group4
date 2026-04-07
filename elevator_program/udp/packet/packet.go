package packet

import (
	"elevator_program/message"
	"encoding/json"
)

type Packet struct {
	Header  Header                  `json:"header"`
	Payload message.ElevatorMessage `json:"payload"`
}

func (p Packet) EncodePacket() ([]byte, error) {
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

func IsBroadcastPkt(t PacketType) bool {
	switch t {
	case PKT_T_BroadcastUpdate,
		PKT_T_BroadcastAck,
		PKT_T_BroadcastCommit,
		PKT_T_BroadcastDone:
		return true
	}
	return false
}
