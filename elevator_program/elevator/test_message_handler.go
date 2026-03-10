package elevator

import (
	"elevator_program/elevio"
	"elevator_program/message"
	"elevator_program/udp/server"
	"fmt"
	"time"
)

func (e Elevator) InitMsg() []message.Message {
	msg := []message.Message{
		{
			MsgType: message.MSG_T_TaskAssign,
			Id:      3,
			Task: elevio.ButtonEvent{
				Floor:  2,
				Button: elevio.BT_HallUp,
			},
			BtnStatus: message.ButtonPending,
		},
		{
			MsgType: message.MSG_T_TaskUpdate,
			Id:      3,
			Task: elevio.ButtonEvent{
				Floor:  0,
				Button: elevio.BT_Cab,
			},
			BtnStatus: message.ButtonPending,
		},
		{
			MsgType: message.MSG_T_TaskUpdate,
			Id:      3,
			Task: elevio.ButtonEvent{
				Floor:  1,
				Button: elevio.BT_HallUp,
			},
			BtnStatus: message.ButtonPending,
		},
		{
			MsgType: message.MSG_T_TaskDelegate,
			Id:      1,
			Task: elevio.ButtonEvent{
				Floor:  1,
				Button: elevio.BT_Cab,
			},
			BtnStatus: message.ButtonPending,
		},
		{
			MsgType:     message.MSG_T_StatusReport,
			Id:          20,
			CabRequests: make([]message.ButtonStatus, 4),
		},
	}

	return msg
}
func (e *Elevator) TestMsgHandler(numFloors int) {
	e.system.Elevators[2] = ElevatorsStatus{
		Id:          2,
		CabRequests: make([]ButtonStatus, numFloors),
	}

	fmt.Println("Initial system state:")
	fmt.Println(e.system)

	// Create test message
	msg := e.InitMsg()

	for id, currMsg := range msg {
		fmt.Println("Msg Id: ", id)
		fmt.Println("\nSending message:", currMsg.MsgType)

		e.MessageHandler(currMsg)

		fmt.Println("\nSystem state after message:")
		fmt.Println(e.system)
		time.Sleep(2 * time.Second)
	}

	// TODO Figure this out need Elevator map in message
	// newCopy := copySystem(&e.system)
	// newMsg := message.Message{
	// 	MsgType:        message.MSG_T_NewToChannel,
	// 	Id:       3,
	// 	elevatorStatus: newCopy.Elevators[1],
	// 	fullstate:      &newCopy,
	// }

	e.system.hallRequests = make([][2]ButtonStatus, numFloors)
	e.system.Elevators = make(map[int]ElevatorsStatus)
	e.nextTarget = elevio.ButtonEvent{
		Floor:  -1,
		Button: elevio.BT_HallUp,
	}
	// elevio.SetMotorDirection(0)
	// e.elevatorState = ES_EmergencyStop
	// fmt.Println("\nEmpty system:")
	// fmt.Println(e.system)
	// time.Sleep(2 * time.Second)
	// e.MessageHandler(newMsg)
	// fmt.Println("\nSystem state after message:")
	// fmt.Println(e.system)
}

func copySystem(original *System) System {
	// Create a new instance of System and copy the fields
	newCopy := *original

	// Deep copy the map to ensure independence
	newCopy.Elevators = make(map[int]ElevatorsStatus)
	for id, elevator := range original.Elevators {
		newCopy.Elevators[id] = elevator
	}

	// Deep copy the hallRequests slice
	newCopy.hallRequests = make([][2]ButtonStatus, len(original.hallRequests))
	copy(newCopy.hallRequests, original.hallRequests)

	return newCopy
}

func (e *Elevator) TestMsgHandler_Master(numFloors int) {

	// Create protocol
	// protocol := Protocol{
	// 	ackArray: make(map[int]int),
	// }

	// Create system
	system := System{
		hallRequests: make([][2]ButtonStatus, numFloors),
		Elevators:    make(map[int]ElevatorsStatus),
	}
	system.Elevators[1] = e.system.Elevators[1]

	// Create master elevator
	e.isMaster = true

	// Create slave elevator
	slave := Elevator{
		// id:       2,
		// isMaster: false,
		system: system,
		// protocol: &protocol,
	}
	slave.system.Elevators[2] = ElevatorsStatus{
		CabRequests: make([]ButtonStatus, numFloors),
		Id:          2,
	}

	slave.system.Elevators[2].CabRequests[2] = Pending

	// TODO Figure out Elevator map
	// fmt.Println("---- TEST: Status Report ----")
	// fmt.Println("System before msg: ", e.system.Elevators)

	// msg := message.Message{
	// 	MsgType:        message.MSG_T_StatusReport,
	// 	Id:       2,
	// 	Ip: "Jeg aner ikke",
	// 	CurrentFloor: 1,
	// 	CabRequests: make([]message.ButtonStatus, 4),
	// 	elevatorStatus: slave.system.Elevators[2],
	// }

	// e.MessageHandler(msg)

	// fmt.Println("System elevator map:", e.system.Elevators)

	fmt.Println("---- TEST: Task Update ----")

	taskMsg := message.Message{
		MsgType: message.MSG_T_TaskUpdate,
		Id:      2,
		Task: elevio.ButtonEvent{
			Floor:  0,
			Button: elevio.BT_HallUp,
		},
		BtnStatus: message.ButtonPending,
	}

	e.MessageHandler(taskMsg)

	fmt.Println("Hall requests:", e.system.hallRequests)

	// fmt.Println("---- TEST: Slave receiving TaskAssign ----")

	// assignMsg := Message{
	// 	msgType: MSG_T_TaskAssign,
	// 	task: elevio.ButtonEvent{
	// 		Floor:  3,
	// 		Button: elevio.BT_HallDown,
	// 	},
	// 	btnStatus: Running,
	// }

	// slave.MessageHandler(assignMsg)

	// fmt.Println("Slave next target:", slave.nextTarget)
}

// call by using e.server as the parameter
// NB need to make two elevators on two different ports, also both "e.server.Listen()"
// must be called as go-routines for them to start listening
func test_sending(srv *server.Server, sessionID uint32, ip_receiver string, port_receiver int) {
	srv.SendSession(sessionID, ip_receiver, port_receiver, message.Message{Ip: "Dette er en test :)"})
}
