package types

import (
	"elevator_program/elevio"
)

// TODO rename
type ElevatorsStatus struct {
	Id           string
	Ip           string
	CurrentFloor int
	CabRequests  []ButtonStatus
	Target       elevio.ButtonEvent
	Direction    elevio.MotorDirection
	State        ElevatorState
	IsMaster     bool
}

// func (e *Elevator) TestMasterLogic() {
// 	fmt.Print("Starting \n")

// 	// ----------------------------
// 	// Setup elevator registry
// 	// ----------------------------

// 	elevatorRegistry := make(map[int]*ElevatorsStatus)

// 	// Elevator 1: Idle at floor 0
// 	elevatorRegistry[1] = &ElevatorsStatus{
// 		State:        ES_Idle,
// 		CurrentFloor: 0,
// 		Target:       elevio.ButtonEvent{},
// 		Direction:    elevio.MD_Stop,
// 	}

// 	// Elevator 2: Moving Up from floor 1 to 3
// 	elevatorRegistry[2] = &ElevatorsStatus{
// 		State:        ES_Moving,
// 		CurrentFloor: 1,
// 		Target:       elevio.ButtonEvent{Floor: 3, Button: elevio.BT_HallUp},
// 		Direction:    elevio.MD_Up,
// 	}

// 	// Elevator 3: Moving Down from floor 3 to 0
// 	elevatorRegistry[3] = &ElevatorsStatus{
// 		State:        ES_Moving,
// 		CurrentFloor: 3,
// 		Target:       elevio.ButtonEvent{Floor: 0, Button: elevio.BT_HallUp},
// 		Direction:    elevio.MD_Down,
// 	}

// 	// ----------------------------
// 	// Create master elevator
// 	// ----------------------------

// 	// master := Elevator{
// 	// 	hallRequests: make([]elevio.ButtonEvent, 0),
// 	// }

// 	// ----------------------------
// 	// Test Cases
// 	// ----------------------------

// 	testCases := []elevio.ButtonEvent{
// 		{Floor: 2, Button: elevio.BT_HallUp},
// 		{Floor: 0, Button: elevio.BT_HallUp},
// 		{Floor: 3, Button: elevio.BT_HallDown},
// 		{Floor: 1, Button: elevio.BT_Cab},
// 		{Floor: 2, Button: elevio.BT_HallUp},
// 		{Floor: 1, Button: elevio.BT_Cab},
// 	}

// 	for _, request := range testCases {

// 		fmt.Println("===================================")
// 		fmt.Printf("New Request: Floor %d, Button %v\n", request.Floor, request.Button)

// 		bestID, distance, _ := e.ClosestToTarget(elevatorRegistry, request)

// 		if bestID == -1 {
// 			fmt.Println("No elevator can take the request")
// 		} else {
// 			fmt.Printf("Assigned to Elevator %d (distance = %d)\n", bestID, distance)
// 		}

// 		// Temp need to change elevator 1 to moving to test other cases
// 		elevatorRegistry[1].State = ES_Moving
// 	}
// 	fmt.Print("Done")
// }
