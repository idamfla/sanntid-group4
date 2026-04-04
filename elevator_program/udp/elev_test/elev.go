package elevtest

import (
	"elevator_program/message"
	"elevator_program/udp"
	"elevator_program/udp/packet"
	"elevator_program/udp/server"
	"elevator_program/udp/session"
	"fmt"
	"net"
	"sync"
)

type Elev struct {
	ID       string
	isMaster bool
	ch       chan session.ElevatorPacket
	srv      *server.Server
	wg       sync.WaitGroup
	stop     chan struct{}
}

func NewElev(id string) *Elev {
	return &Elev{
		ID:   id,
		ch:   make(chan session.ElevatorPacket, 32),
		stop: make(chan struct{}),
	}
}

func (e *Elev) StartServer(ip string, port int) error {
	srv, err := server.NewServer(ip, port, e.ID, e.ch)
	if err != nil {
		return err
	}

	e.srv = srv
	fmt.Println("Server", e.srv.ID, "is running ...")
	return nil
}

func (e *Elev) listen() {
	defer e.wg.Done()
	for {
		select {
		case msg := <-e.ch:
			fmt.Println("elev got elevator packet:", msg.EMsg)
			if msg.Done != nil {
				msg.Done <- struct{}{}
			}

			if msg.EMsg.EMsgType == message.EMSG_T_NewToChannel {
				fmt.Println(e.srv.ID, "took snapshot")
				udpAddr, err := udp.StringAddrToUDPAddr(msg.EMsg.Addr)
				if err != nil {
					continue
				}
				e.srv.QueueMessage(udpAddr, packet.PROTO_PKT_T_Snapshot, message.ElevatorMessage{})
				e.srv.StartPeerCatchup(udpAddr)
			}
		case <-e.stop:
			return
		}
	}
}

func (e *Elev) Start() {
	e.wg.Add(1)
	go e.listen()
	go e.srv.Start()
}

func (e *Elev) QueueMessage(remoteAddr *net.UDPAddr, protoPktType packet.ProtocolPacketType, eMsg message.ElevatorMessage) {
	e.srv.QueueMessage(remoteAddr, protoPktType, eMsg)

}

func (e *Elev) Close() {
	close(e.stop)

	if e.srv != nil {
		e.srv.PrintSessions()
		e.wg.Add(1)
		go func() {
			defer e.wg.Done()
			e.srv.Close()
		}()
	}

	e.wg.Wait()

	close(e.ch)

	fmt.Printf("Elevator %s and server have shut down cleanly\n", e.ID)
}

func (e *Elev) IsMaster() bool {
	return e.srv.IsMaster()
}
