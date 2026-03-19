package coordinator

import (
	"elevator_program/elevator"
	"elevator_program/message"
	"elevator_program/udp/packet"
	"elevator_program/udp/server"
	"elevator_program/udp/session"
	"fmt"
	"net"
	"sync"
	"time"
)

// Delegating tasks to own elevator and preparing and queueing messages to be sent over udp
type Coordinator struct {
	Server        *server.Server
	msgRecieveCh  chan session.ElevatorPacket
	msgSendCh     chan message.ElevatorMessage
	wg            sync.WaitGroup
	portRegistery map[string]int
	TaskMonitor   TaskMonitor
}

// Initialize the coordinator struct
func (c *Coordinator) InitCoordinator() {
	c.msgRecieveCh = make(chan session.ElevatorPacket, 10)
	c.msgSendCh = make(chan message.ElevatorMessage, 10)
	c.portRegistery = map[string]int{
		"broadcast": 3000, // TODO change this
		"master":    9000,
	}
	c.TaskMonitor = NewTaskMonitor(15 * time.Second) // TODO how long should we wait??
}

// For starting server
func (c *Coordinator) StartServer(ip string, port int, id string) error {
	srv, err := server.NewServer(ip, port, id, c.msgRecieveCh)
	if err != nil {
		return err
	}

	c.Server = srv
	fmt.Println("Server", c.Server.ID, "is running...")
	return nil
}

// Start necessary gorutines for the coordinator
func (c *Coordinator) Start(e *elevator.Elevator) {
	c.wg.Add(1)
	go c.MessageListener(e)
	go c.sendListener(e)
	go c.Server.Start()
}

// For queueing messages to send with udp
func (c *Coordinator) QueueMessage(remoteAddr *net.UDPAddr, protoPktType packet.ProtocolPacketType, msg message.ElevatorMessage) {
	c.Server.QueueMessage(remoteAddr, protoPktType, msg)
}

// Closes the server
func (c *Coordinator) Close() {
	close(c.msgRecieveCh)

	if c.Server != nil {
		c.Server.PrintSessions()
		c.Server.Close()
	}

	fmt.Println("Elevator and server have shut down cleanly")
}
