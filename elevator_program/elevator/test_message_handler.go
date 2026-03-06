package elevator

import (
	"elevator_program/elevio"
	"fmt"
	"time"
)

func (e Elevator) InitMsg() []Message {
	msg := []Message{
		{
			msgType:  MSG_T_TaskAssign,
			senderId: 3,
			task: elevio.ButtonEvent{
				Floor:  2,
				Button: elevio.BT_HallUp,
			},
			btnStatus: Pending,
		},
		{
			msgType:  MSG_T_TaskUpdate,
			senderId: 3,
			task: elevio.ButtonEvent{
				Floor:  0,
				Button: elevio.BT_Cab,
			},
			btnStatus: Pending,
		},
		{
			msgType:  MSG_T_TaskUpdate,
			senderId: 3,
			task: elevio.ButtonEvent{
				Floor:  1,
				Button: elevio.BT_HallUp,
			},
			btnStatus: Pending,
		},
		{
			msgType:  MSG_T_TaskDelegate,
			senderId: 3,
			task: elevio.ButtonEvent{
				Floor:  1,
				Button: elevio.BT_Cab,
			},
			btnStatus:           Pending,
			idToElevatorMission: 1,
		},
		{
			msgType:  MSG_T_StatusReport,
			senderId: 2,
			elevatorStatus: ElevatorsStatus{
				id:          20,
				cabRequests: make([]ButtonStatus, 4),
			},
		},
	}

	return msg
}
func (e *Elevator) TestMsgHandler(numFloors int) {
	e.system.Elevators[2] = ElevatorsStatus{
		id:          2,
		cabRequests: make([]ButtonStatus, numFloors),
	}

	fmt.Println("Initial system state:")
	fmt.Println(e.system)

	// Create test message
	msg := e.InitMsg()

	for id, currMsg := range msg {
		fmt.Println("Msg Id: ", id)
		fmt.Println("\nSending message:", currMsg.msgType)

		e.MessageHandler(currMsg)

		fmt.Println("\nSystem state after message:")
		fmt.Println(e.system)
		time.Sleep(2 * time.Second)
	}

	newCopy := copySystem(&e.system)
	newMsg := Message{
		msgType:        MSG_T_NewToChannel,
		senderId:       3,
		elevatorStatus: newCopy.Elevators[1],
		fullstate:      &newCopy,
	}

	// fmt.Println("Riktig data: ", newMsg.fullstate.hallRequests)
	// fmt.Println("Riktig data Heiser: ", newMsg.fullstate.Elevators)
	e.system.hallRequests = make([][2]ButtonStatus, numFloors)
	e.system.Elevators = make(map[int]ElevatorsStatus)
	e.nextTarget = elevio.ButtonEvent{
		Floor:  -1,
		Button: elevio.BT_HallUp,
	}
	// fmt.Println("Fortsatt riktig data?: ", newMsg.fullstate.hallRequests)
	// fmt.Println("Fortsatt riktig data heiser?: ", newMsg.fullstate.Elevators)
	elevio.SetMotorDirection(0)
	e.elevatorState = ES_EmergencyStop
	fmt.Println("\nEmpty system:")
	fmt.Println(e.system)
	time.Sleep(2 * time.Second)
	e.MessageHandler(newMsg)
	fmt.Println("\nSystem state after message:")
	fmt.Println(e.system)
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
		cabRequests: make([]ButtonStatus, numFloors),
		id:          2,
	}

	slave.system.Elevators[2].cabRequests[2] = Pending

	fmt.Println("---- TEST: Status Report ----")
	fmt.Println("System before msg: ", e.system.Elevators)

	msg := Message{
		msgType:        MSG_T_StatusReport,
		senderId:       2,
		elevatorStatus: slave.system.Elevators[2],
	}

	e.MessageHandler(msg)

	fmt.Println("System elevator map:", e.system.Elevators)

	fmt.Println("---- TEST: Task Update ----")

	taskMsg := Message{
		msgType:  MSG_T_TaskUpdate,
		senderId: 2,
		task: elevio.ButtonEvent{
			Floor:  0,
			Button: elevio.BT_HallUp,
		},
		btnStatus: Pending,
		comNumber: 1,
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
