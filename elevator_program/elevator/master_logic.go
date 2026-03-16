package elevator
import (
	"fmt"
	"time"

)
func (e *Elevator) InitMasterElevator() {
    fmt.Printf("Elevator %d initializing master role\n", e.id)
	e.isMaster = true
	e.currentMasterID = e.id
	e.connectedToMaster = true

	// boradcast "I am master, here is my recieve ch"
	// broadcast latest system state to everyone

}

func (e *Elevator) RunMasterLoop() {
    ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for range ticker.C {
		if !e.isMaster {
			return
		}

    }
}

// iterate over map of name elevators
// for _, e := range elevators {
//     if e.State == ES_Idle {
//         ...
//     }
// }
