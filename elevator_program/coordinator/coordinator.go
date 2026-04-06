package coordinator

import (
	"elevator_program/elevator"
	"elevator_program/message"
	"elevator_program/udp/packet"
	"elevator_program/udp/server"
	"elevator_program/udp/session"
	"fmt"
	"sync"
	"time"
)

type Coordinator struct {
	Server        *server.Server
	msgRecieveCh  chan session.ElevatorPacket
	msgSendCh     chan message.ElevatorMessage
	portRegistery map[string]int
	TaskMonitor   TaskMonitor

	wg   sync.WaitGroup
	stop chan struct{}
}

func (c *Coordinator) InitCoordinator() {
	c.msgRecieveCh = make(chan session.ElevatorPacket, 20)
	c.msgSendCh = make(chan message.ElevatorMessage, 20)
	c.TaskMonitor = NewTaskMonitor(10 * time.Second)
}

func (c *Coordinator) StartServer(ip string, port int, id string) error {
	srv, err := server.NewServer(ip, port, id, c.msgRecieveCh)
	if err != nil {
		return err
	}

	c.Server = srv
	fmt.Println("Server", c.Server.Alias, "is running...")
	return nil
}

func (c *Coordinator) Start(e *elevator.Elevator) {
	c.wg.Add(1)
	go c.MessageListener(e)
	go c.sendListener(e)
	go c.Server.Start()
	// go c.stateMonitor(e)
}

func (c *Coordinator) QueueMessage(protoPktType packet.ProtocolPacketType, msg message.ElevatorMessage) {
	c.Server.QueueMessage(protoPktType, msg)
}

func (c *Coordinator) Close() {

	if c.Server != nil {
		c.Server.PrintSessions()
		c.Server.Close()
	}

	close(c.msgRecieveCh)
	fmt.Println("Elevator and server have shut down cleanly")
}
