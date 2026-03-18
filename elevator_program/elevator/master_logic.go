package elevator

import (
	"elevator_program/message"
	"elevator_program/types"
	"fmt"
	"time"
)

func (e *Elevator) InitMasterElevator() {
	fmt.Printf("Elevator %s initializing master role\n", e.Id)
	e.IsMaster = true
	e.currentMasterID = e.Id
	e.connectedToMaster = true

	// boradcast "I am master, here is my recieve ch"
	// broadcast latest system state to everyone

}

func (e *Elevator) RunMasterLoop() {
	ticker := time.NewTicker(300 * time.Millisecond)
	defer ticker.Stop()

	for range ticker.C {
		if !e.IsMaster {
			return
		}

		msg := message.ElevatorMessage{
			MsgType: types.MSG_T_StatusReport,
			Id:      e.Id,
			Elevators: map[string]types.ElevatorsStatus{
				e.Id: e.System.Elevators[e.Id],
			},
		}

		e.SendToCoordinator <- msg
	}
}

// iterate over map of name elevators
// for _, e := range elevators {
//     if e.State == ES_Idle {
//         ...
//     }
// }
