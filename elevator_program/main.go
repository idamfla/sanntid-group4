
package main

import (
	"elevator_program/config"
	"elevator_program/coordinator"
	"elevator_program/elevator"
	"elevator_program/elevio"
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"syscall"
)

const localIP = "127.0.0.1"

func runOne(cfg config.Config) {
	addr := fmt.Sprintf("%s:%s", cfg.IP, cfg.Port)
	fmt.Println("Connecting to elevator at", addr)

	elevio.Init(addr, cfg.Floors)
	fmt.Println("elevio.Init finished")

	var e elevator.Elevator
	e.InitElevator(strconv.Itoa(cfg.ID), cfg.Floors, cfg.InitFloor, localIP, 9000+(cfg.ID-1))

	var c coordinator.Coordinator
	c.InitProtocol()

	err := c.StartServer(localIP, 9000+(cfg.ID-1), strconv.Itoa(cfg.ID))
	if err != nil {
		fmt.Println("Could not start server:", err)
		return
	}

	c.Start(&e)
	e.RunElevatorProgram()

	fmt.Printf("Started elevator %d\n", cfg.ID)
	fmt.Printf("Simulator address: %s\n", addr)
	fmt.Printf("Coordinator server: %s:%d\n", localIP, 9000+(cfg.ID-1))



	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	<-sigChan

	fmt.Println("\nShutting down...")
	c.Close()
}

func main() {
	cfg := config.ParseFlags()

	if cfg.ID == 0 {
		fmt.Println("Launcher mode")
		config.SpawnElevators(cfg)
		return
	}

	runOne(cfg)
}


/*
package main

import (
	"elevator_program/coordinator"
	"elevator_program/config"
	"elevator_program/elevator"
	"elevator_program/message"
	"elevator_program/protocol"
	"elevtest "elevator_program/udp/elev_test"
	"elevator_program/udp/packet"
	"elevator_program/udp/server"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	// "elevator_program/utilities"
	"elevator_program/elevio"
)

const (
	localIP  = "127.0.0.1"
	receiver = "10.100.23.15"
)

func testElevator() {
	var e elevator.Elevator
	// fmt.Println(e)

	id := "A"
	numFloors := 4
	initFloor := 3 // NB! in the code the elevator floors are 0-index, on the controller it is not
	ip_address := "localhost"
	port := "15657"

	// "localhost:15657"
	elevio.Init(ip_address+":"+port, numFloors)

	e.InitElevator(id, numFloors, initFloor, ip_address, 5000) // TODO WHAT TO DO HERE, prot is int??
	e.RunElevatorProgram()

	p := coordinator.Coordinator{}
	p.InitProtocol()
	// p.StartServer(ip_address, 5000, id)
	p.Start(&e)
	// TODO MAybe the right spot to put it
	defer p.Close()

	// go p.TestMsgHandler(&e, numFloors)
	// go p.TestMsgHandler_Master(&e, numFloors)

	// e.TestMasterLogic()

	/*
		TODO, bug - when cab to floor 2, then cab to floor 1, if floor 3 is pressed after reaching floor 2, elevator will go up to floor 3

	select {}
}

func testBroadcast_send(srv *server.Server) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	bcMsg := message.Message{Ip: "Hello, broadcast from " + srv.ID}

	for range ticker.C {
		srv.QueueMessage(nil, packet.PROTO_PKT_T_BroadcastUpdate, bcMsg)
		fmt.Println("bcMsg:", srv.ID, ",", bcMsg)
	}
}

// just to get some prints after i "shut down"
func closeProgram(e1 *elevtest.Elev, e2 *elevtest.Elev) {
	// Create signal channel
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	// Wait for Ctrl+C
	<-sigChan

	fmt.Println("\nCtrl+C pressed")

	// Graceful shutdown
	e1.Close()
	e2.Close()

	fmt.Println("Servers shut down cleanly")
}

func main() {
	// eA := elevtest.NewElev("A", 2)

	// err := eA.StartServer(localIP, 9000)
	// if err != nil {
	// 	fmt.Println(err)
	// 	return
	// }

	// eB := elevtest.NewElev("B", 2)

	// err = eB.StartServer(localIP, 9001)
	// if err != nil {
	// 	fmt.Println(err)
	// 	return
	// }

	// eA.Start()
	// eB.Start()

	// eA.QueueMessage(
	// 	udp.MustUDPAddr(localIP, 9001),
	// 	packet.PROTO_PKT_T_BroadcastData,
	// 	message.Message{Ip: "Hello A!"},
	// )

	// eA.QueueMessage(
	// 	udp.MustUDPAddr(localIP, 9001),
	// 	packet.PROTO_PKT_T_BroadcastData,
	// 	message.Message{Ip: "Hello Ax2!"},
	// )
	// closeProgram(eA, eB)

	ip_address := "localhost"
	port := "15657"

	elevio.Init(ip_address+":"+port, 4)

	e1 := elevator.Elevator{}
	e1.InitElevator("1", 4, 3, localIP, 9000)
	p1 := coordinator.Coordinator{}
	p1.InitProtocol()
	err := p1.StartServer(localIP, 9000, "1")
	if err != nil {
		fmt.Println(err)
		return
	}
	p1.Start(&e1)

	// fmt.Println("eeeee: ", e1)

	fmt.Println("Created e1 and p1")

	e2 := elevator.Elevator{}
	e2.InitElevator("2", 4, 3, localIP, 9001)
	p2 := coordinator.Coordinator{}
	p2.InitProtocol()
	err = p2.StartServer(localIP, 9001, "2")
	if err != nil {
		fmt.Println(err)
		return
	}
	// e2copy := e2.System.Elevators[e2.Id]
	// e2copy.CabRequests[2] = types.Running
	// e2.System.Elevators[e2.Id] = e2copy
	// fmt.Println("Created e2 and p2")

	p2.Start(&e2)

	defer p1.Close()
	defer p2.Close()

	fmt.Println("finished init")
	// e1.IsOnline = true
	e2.RunElevatorProgram()

	// vierdElev := types.ElevatorsStatus{
	// 	Ip:          "Halla balla",
	// 	CabRequests: make([]types.ButtonStatus, 4),
	// }

	// msg := message.Message{
	// 	MsgType: types.MSG_T_NewToChannel,
	// 	Id:      "2",
	// 	Elevators: map[string]types.ElevatorsStatus{
	// 		"2": vierdElev,
	// 	},
	// }
	// time.Sleep(2 * time.Second)

	// msg := message.Message{
	// 	MsgType: types.MSG_T_LostComs,
	// 	Id:      "1",
	// 	Ip:      e1.Ip,
	// }

	// fmt.Println("Connected to master, ", e2.ConnectedToMaster())
	// e1.IsMaster = false
	// time.Sleep(2 * time.Second)
	// fmt.Println("Sending now")
	// p1.SendMessageSlave(&e1, msg)

	// msg = message.Message{
	// 	MsgType:   types.MSG_T_TaskUpdate,
	// 	Id:        "1",
	// 	BtnStatus: types.Pending,
	// 	Task: elevio.ButtonEvent{
	// 		Floor:  1,
	// 		Button: elevio.BT_HallUp,
	// 	},
	// }
	time.Sleep(2 * time.Second)

	fmt.Println("Is it now connected to system? ", e1.Id, e1.IsMaster, e1.IsOnline)
	// fmt.Println("Sending now")
	// p2.SendMessageMaster(msg)

	select {}

	//TODO make sure the server dosent take in peers and not totalnumberofElevators
}

// 127.0.0.1 er lokal <- du kan bruke denne for en til en
// 10.22.67.255 broadcast, broadcast har 255 som siste verdi

*/
