package elevtest

import (
	"elevator_program/udp/message"
	"elevator_program/udp/packet"
	"elevator_program/udp/server"
	"elevator_program/udp/session"
	"fmt"
	"net"
	"sync"
)

type Elev struct {
	ID  string
	ch  chan session.ElevatorPacket
	srv *server.Server
	wg  sync.WaitGroup
}

func NewElev(id string) *Elev {
	return &Elev{
		ID: id,
		ch: make(chan session.ElevatorPacket),
	}
}

func (e *Elev) StartServer(ip string, port int) error {
	srv, err := server.NewServer(ip, port, e.ID, e.ch)
	if err != nil {
		return err
	}

	e.srv = srv
	fmt.Println("Server", e.srv.ID, "is running...")
	return nil
}

func (e *Elev) listen() {
	defer e.wg.Done()
	for msg := range e.ch {
		fmt.Println("Got elevator packet:", msg.Packet.Payload.Content)
		if msg.Done != nil {
			msg.Done <- struct{}{}
		}
	}
}

func (e *Elev) Start() {
	e.wg.Add(1)
	go e.listen()
	go e.srv.Start()
}

func (e *Elev) QueueMessage(remoteAddr *net.UDPAddr, protoPktType packet.ProtocolPacketType, msg message.Message) {
	e.srv.QueueMessage(remoteAddr, protoPktType, msg)

}

func (e *Elev) Close() {
	close(e.ch)

	e.srv.PrintSessions()
	e.srv.Close()

	fmt.Printf("Elevator %s and server have shut down cleanly\n", e.ID)
}
