package elevator

import (
	"fmt"
	"time"

	// "elevator_program/utilities"
	"elevator_program/elevio"
)

type Elevator struct {
	id int

	inBetweenFloors bool
	currentFloor    int
	nextTarget      elevio.ButtonEvent
	direction       elevio.MotorDirection
	initFloor       int

	hallRequests [][2]bool // TODO maybe Pending, Running, Completed, NotActive
	cabRequests  []bool

	doorState DoorState
	doorTimer time.Time

	elevatorState    ElevatorState
	obstruction      bool
	emergencyStop    bool // TODO fade out ... just figure out how to set state to ES_EmergencyStop, unset it
	hardwareEventsCh chan HardwareEvent

	// msgRecieveCh chan Message
	// msgSendCh    chan Message

	isMaster         bool
	elevatorRegistry map[string]ElevatorsStatus
}

func (e *Elevator) InitElevator(id int, numFloors int, initFloor int) {
	e.id = id
	e.currentFloor = -1
	e.nextTarget = elevio.ButtonEvent{Floor: -1}
	e.initFloor = initFloor
	e.doorTimer = time.Time{}
	e.hallRequests = make([][2]bool, numFloors)
	e.cabRequests = make([]bool, numFloors)
	// e.elevatorState = ES_Uninitialized

	e.hardwareEventsCh = make(chan HardwareEvent, 20)

	// e.StatusChan = statusChan
	// e.TaskChan = taskChan

	// e.StatusChan <-utilities.StatusMsg{e.id, e.currentFloor, e.nextTarget}

	e.clearAllLamps(elevio.BT_HallUp, elevio.BT_HallDown, elevio.BT_Cab)
}

func (e *Elevator) RunElevatorProgram() {
	fmt.Println("RUNNING ELEVATOR PROGRAM")
	go e.RunHardwareEventLoop()
	go e.RunDoorStateMachine()
	go e.RunElevatorStateMachine()
	e.StartHardwareEventsListeners()

	done := make(chan struct{})
	<-done
}

// region printing, for debugging
func (e Elevator) String() string {
	s := fmt.Sprintf(
		`Elevator
	id: %d
	in between floors: %t
	current floor: %d
	target: %d, %s
	init floor: %d
	last moving dir: %s
	door state: %s
	elevator state: %s
`,
		e.id, e.inBetweenFloors, e.currentFloor, e.nextTarget.Floor, e.nextTarget.Button, e.initFloor, e.direction, e.doorState, e.elevatorState)

	for f := 0; f < len(e.hallRequests); f++ {
		req := e.hallRequests[f]
		cab := e.cabRequests[f]

		s += fmt.Sprintf(
			"	floor %d: [Up:%t Down:%t Cab:%t]\n",
			f,
			req[elevio.BT_HallUp],
			req[elevio.BT_HallDown],
			cab,
		)
	}

	return s
}

// endregion
