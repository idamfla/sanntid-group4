package elevator

import (
	"fmt"
	"time"

	"elevator_program/elevio"
	"elevator_program/message"
	"elevator_program/system"
	"elevator_program/types"
)

type Elevator struct {
	Id int
	// TODO temp need to know the ip using the id
	IpRegistery map[string]int

	inBetweenFloors bool
	currentFloor    int
	nextTarget      elevio.ButtonEvent
	direction       elevio.MotorDirection
	initFloor       int

	hallRequests [][2]types.ButtonStatus // TODO Should remove these
	cabRequests  []types.ButtonStatus    // TODO Should remove these

	doorState DoorState
	doorTimer time.Time

	elevatorState    types.ElevatorState
	obstruction      bool
	emergencyStop    bool // TODO fade out ... just figure out how to set state to ES_EmergencyStop, unset it
	hardwareEventsCh chan HardwareEvent

	MsgRecieveCh chan message.Message
	msgSendCh    chan message.Message

	IsMaster          bool
	connectedToMaster bool
	// elevatorRegistry  map[int]types.ElevatorsStatus // TODO Was string, could also make it uint

	// TODO Trying to split ut the code
	System system.System

	// server server.Server // TODO Something here makes everything yellow, complaining about locks
}

func (e *Elevator) InitElevator(id int, numFloors int, initFloor int, ip string, port string) { // TODO Changed port to string, hope everything works
	e.Id = id
	e.currentFloor = -1
	e.nextTarget = elevio.ButtonEvent{Floor: -1}
	e.initFloor = initFloor
	e.doorTimer = time.Time{}
	e.hallRequests = make([][2]types.ButtonStatus, numFloors)
	e.cabRequests = make([]types.ButtonStatus, numFloors)

	e.IsMaster = false

	e.System.InitSystem(1, "192.168.0.1", 4)

	e.elevatorState = types.ES_Uninitialized

	// TODO Need to initialize system

	e.hardwareEventsCh = make(chan HardwareEvent, 20)

	// e.StatusChan = statusChan
	// e.TaskChan = taskChan

	// e.StatusChan <-utilities.StatusMsg{e.id, e.currentFloor, e.nextTarget}
	e.MsgRecieveCh = make(chan message.Message, 10) // TODO Should have this in the code
	// server = server.NewServer(ip, port, e.id, msgRecieveCh) // TODO should have this in the qode

	e.clearAllLamps(elevio.BT_HallUp, elevio.BT_HallDown, elevio.BT_Cab)
}

func (e *Elevator) RunElevatorProgram() {
	fmt.Println("RUNNING ELEVATOR PROGRAM")
	go e.RunHardwareEventLoop()
	go e.RunDoorStateMachine()
	go e.RunElevatorStateMachine()
	e.StartHardwareEventsListeners()
	time.Sleep(10 * time.Second)

	// Temp for testing msgHandler
	// go e.TestMsgHandler(4)
	// go e.TestMsgHandler_Master(4)
	// go e.server.Listen() // TODO check that this work
	// done := make(chan struct{})
	// <-done
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
		e.Id, e.inBetweenFloors, e.currentFloor, e.nextTarget.Floor, e.nextTarget.Button, e.initFloor, e.direction, e.doorState, e.elevatorState)

	for f := 0; f < len(e.hallRequests); f++ {
		req := e.System.HallRequests[f]
		cab := e.System.Elevators[e.Id].CabRequests[f]

		s += fmt.Sprintf(
			"	floor %d: [Up:%s Down:%s Cab:%s]\n",
			f,
			req[elevio.BT_HallUp],
			req[elevio.BT_HallDown],
			cab,
		)
	}

	return s
}

// endregion
