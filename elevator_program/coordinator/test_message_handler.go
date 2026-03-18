package coordinator

// import (
// 	"elevator_program/elevator"
// 	"elevator_program/elevio"
// 	"elevator_program/message"
// 	"elevator_program/system"
// 	"elevator_program/types"
// 	"elevator_program/udp/packet"
// 	"elevator_program/udp/session"
// 	"fmt"
// 	"time"
// )

// func (c *Coordinator) InitMsg() []message.ElevatorMessage {
// 	msg := []message.ElevatorMessage{
// 		{
// 			MsgType: types.MSG_T_TaskUpdate,
// 			Id:      "1",
// 			Task: elevio.ButtonEvent{
// 				Floor:  1,
// 				Button: elevio.BT_HallUp,
// 			},
// 			BtnStatus: types.Running,
// 		},
// 		{
// 			MsgType: types.MSG_T_TaskUpdate,
// 			Id:      "1",
// 			Task: elevio.ButtonEvent{
// 				Floor:  0,
// 				Button: elevio.BT_Cab,
// 			},
// 			BtnStatus: types.Running,
// 		},
// 		{
// 			MsgType: types.MSG_T_TaskUpdate,
// 			Id:      "3",
// 			Task: elevio.ButtonEvent{
// 				Floor:  1,
// 				Button: elevio.BT_HallUp,
// 			},
// 			BtnStatus: types.Pending,
// 		},
// 		{
// 			MsgType: types.MSG_T_TaskUpdate,
// 			Id:      "2",
// 			Task: elevio.ButtonEvent{
// 				Floor:  1,
// 				Button: elevio.BT_Cab,
// 			},
// 			BtnStatus: types.Pending,
// 		},
// 		// {
// 		// 	MsgType: types.MSG_T_StatusReport,
// 		// 	Id:      "2",
// 		// 	Elevators: map[string]types.ElevatorsStatus{
// 		// 		"2": types.ElevatorsStatus{
// 		// 			Id:          "20",
// 		// 			CabRequests: []types.ButtonStatus{types.Pending, types.NotActive, types.NotActive, types.Running},
// 		// 		},
// 		// 	},
// 		// },
// 	}

// 	return msg
// }

// func (c *Coordinator) TestMsgHandler(e *elevator.Elevator, numFloors int) {
// 	e.System.Elevators["2"] = types.ElevatorsStatus{
// 		Id:          "2",
// 		CabRequests: make([]types.ButtonStatus, numFloors),
// 	}

// 	fmt.Println("Initial system state:")
// 	fmt.Println(e.System)

// 	// Create test message
// 	msg := c.InitMsg()

// 	for id, currMsg := range msg {
// 		fmt.Println("Msg Id: ", id)
// 		fmt.Println("\nSending message:", currMsg.MsgType)
// 		tempPacket := session.ElevatorPacket{
// 			Packet: packet.Packet{
// 				Payload: currMsg,
// 			},
// 		}

// 		c.msgRecieveCh <- tempPacket
// 		time.Sleep(2 * time.Second)
// 		// p.MessageHandler(e, currMsg)

// 		fmt.Println("\nSystem state after message:")
// 		fmt.Println(e.System)
// 		// time.Sleep(2 * time.Second)
// 	}

// 	// TODO Figure this out need Elevator map in message
// 	newCopy := e.System.CopySystem()
// 	newMsg := message.ElevatorMessage{
// 		MsgType:      types.MSG_T_NewToChannel,
// 		Id:           "1",
// 		Elevators:    newCopy.Elevators,
// 		HallRequests: newCopy.HallRequests,
// 	}

// 	tempPacket := session.ElevatorPacket{
// 		Packet: packet.Packet{
// 			Payload: newMsg,
// 		},
// 	}
// 	// TODO When the elevator has forgot everything, it will crash if a button is preesed
// 	e.ClearElevator(numFloors)
// 	fmt.Println("\nEmpty system:")
// 	fmt.Println(e.System)
// 	time.Sleep(2 * time.Second)
// 	// p.MessageHandler(e, newMsg)
// 	c.msgRecieveCh <- tempPacket
// 	fmt.Println("\nSystem state after message:")
// 	fmt.Println(e.System)
// }

// func (c *Coordinator) TestMsgHandler_Master(e *elevator.Elevator, numFloors int) {

// 	// Create system
// 	system := system.System{
// 		HallRequests: make([][2]types.ButtonStatus, numFloors),
// 		Elevators:    make(map[string]types.ElevatorsStatus),
// 	}
// 	system.Elevators["1"] = e.System.Elevators["1"]

// 	// Create master elevator
// 	e.IsMaster = true

// 	// Create slave elevator
// 	slave := e.Create_slave(system)
// 	slave.System.Elevators["2"] = types.ElevatorsStatus{
// 		CabRequests: make([]types.ButtonStatus, numFloors),
// 		Id:          "2",
// 	}

// 	slave.System.Elevators["2"].CabRequests[2] = types.Pending

// 	fmt.Println("---- TEST: Status Report ----")
// 	fmt.Println("System before msg: ", e.System.Elevators)

// 	msg := message.ElevatorMessage{
// 		MsgType: types.MSG_T_StatusReport,
// 		Id:      "2",
// 		Elevators: map[string]types.ElevatorsStatus{
// 			"2": slave.System.Elevators["2"],
// 		},
// 	}

// 	c.MessageHandler(e, msg)

// 	fmt.Println("System elevator map:", e.System.Elevators)

// 	fmt.Println("---- TEST: Task Update ----")

// 	taskMsg := message.ElevatorMessage{
// 		MsgType: types.MSG_T_TaskUpdate,
// 		Id:      "2",
// 		Task: elevio.ButtonEvent{
// 			Floor:  1,
// 			Button: elevio.BT_HallUp,
// 		},
// 		BtnStatus: types.Pending,
// 	}

// 	c.MessageHandler(&slave, taskMsg)

// 	fmt.Println("Hall requests:", slave.System.HallRequests)
// 	fmt.Println("System elevator map:", slave.System.Elevators)

// 	// fmt.Println("---- TEST: Slave receiving TaskAssign ----")

// 	// assignMsg := ElevatorMessage{
// 	// 	msgType: MSG_T_TaskAssign,
// 	// 	task: elevio.ButtonEvent{
// 	// 		Floor:  3,
// 	// 		Button: elevio.BT_HallDown,
// 	// 	},
// 	// 	btnStatus: Running,
// 	// }

// 	// slave.MessageHandler(assignMsg)

// 	// fmt.Println("Slave next target:", slave.nextTarget)
// }

// // TODO Need to work on this
// // call by using e.server as the parameter
// // NB need to make two elevators on two different ports, also both "e.server.Listen()"
// // must be called as go-routines for them to start listening
// // func test_sending(srv *server.Server, sessionID uint32, ip_receiver string, port_receiver int) {
// // 	srv.SendSession(sessionID, ip_receiver, port_receiver, message.ElevatorMessage{Ip: "Dette er en test :)"})
// // }
