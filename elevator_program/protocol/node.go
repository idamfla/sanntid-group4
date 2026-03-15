package protocol

import (
	"elevator_program/elevator"
	"elevator_program/message"
	"elevator_program/udp/packet"
	"elevator_program/udp/server"
	"fmt"
	"net"
)

// TODO change name from protocol

func (p *Protocol) StartServer(ip string, port int, id string, numElevators int) error {
	srv, err := server.NewServer(ip, port, id, numElevators, p.msgRecieveCh)
	if err != nil {
		return err
	}

	p.Server = srv
	fmt.Println("Server", p.Server.ID, "is running...")
	return nil
}

func (p *Protocol) Start(e *elevator.Elevator) {
	p.wg.Add(1)
	go p.MessageListener(e)
	go p.Server.Start()
}

func (p *Protocol) QueueMessage(remoteAddr *net.UDPAddr, protoPktType packet.ProtocolPacketType, msg message.Message) {
	p.Server.QueueMessage(remoteAddr, protoPktType, msg)

}

func (p *Protocol) Close() {
	close(p.msgRecieveCh)

	p.Server.PrintSessions()
	p.Server.Close()

	fmt.Println("Elevator and server have shut down cleanly")
}
