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

type Coordinator struct {
	Server        *server.Server              // TODO be carefull with pass by value functions, locks
	msgRecieveCh  chan session.ElevatorPacket // Update the channel type, wait should this one be IncomingPacket, do i need to debug and encode this one?
	msgSendCh     chan message.ElevatorMessage
	wg            sync.WaitGroup
	activePeers   int
	portRegistery map[string]int
	portSelf      int
	TaskMonitor   TaskMonitor
}

func (c *Coordinator) InitCoordinator() { // TODO how can I allways now excactly how many elevators we are going to use
	c.msgRecieveCh = make(chan session.ElevatorPacket, 10) // Match the expected type
	c.msgSendCh = make(chan message.ElevatorMessage, 10)
	c.activePeers = 1
	c.portRegistery = map[string]int{
		"broadcast": 3000,
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

func (c *Coordinator) Start(e *elevator.Elevator) {
	c.wg.Add(1)
	go c.MessageListener(e)
	go c.sendListener(e)
	go c.Server.Start()
}

func (c *Coordinator) QueueMessage(remoteAddr *net.UDPAddr, protoPktType packet.ProtocolPacketType, msg message.ElevatorMessage) {
	c.Server.QueueMessage(remoteAddr, protoPktType, msg)

}

func (c *Coordinator) Close() {
	close(c.msgRecieveCh)

	c.Server.PrintSessions()
	c.Server.Close()

	fmt.Println("Elevator and server have shut down cleanly")
}
