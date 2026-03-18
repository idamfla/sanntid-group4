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
	Id string
	Ip string // TODO Think we should have this one here
	// TODO temp need to know the ip using the id
	IpRegistery map[string]string

	inBetweenFloors bool
	currentFloor    int
	// nextTarget      elevio.ButtonEvent
	// direction elevio.MotorDirection
	initFloor int

	hallRequests [][2]types.ButtonStatus // TODO Should remove these
	cabRequests  []types.ButtonStatus    // TODO Should remove these

	doorState DoorState
	doorTimer time.Time

	//temp Need to time how long you have lost communiction
	lostComsTimer      time.Time
	ackCounterLostComs int

	// TODO Maybe temp need to notify protocol to send something
	SendToProtocol chan message.ElevatorMessage

	// elevatorState    types.ElevatorState
	obstruction      bool
	emergencyStop    bool // TODO fade out ... just figure out how to set state to ES_EmergencyStop, unset it
	hardwareEventsCh chan HardwareEvent

	// MsgRecieveCh chan message.ElevatorMessage
	// msgSendCh    chan message.ElevatorMessage
	// MsgRecieveCh chan session.ElevatorPacket // Update the channel type, wait should this one be IncomingPacket, do i need to debug and encode this one?

	IsMaster          bool
	connectedToMaster bool
	IsOnline          bool

	System system.System

	// Server *server.Server // TODO be carefull with pass by value functions, locks
}

func (e *Elevator) InitElevator(id string, numFloors int, initFloor int, ip string, port int) { // TODO Changed port to string, hope everything works
	e.Id = id
	e.currentFloor = -1
	// e.nextTarget = elevio.ButtonEvent{Floor: -1}
	e.initFloor = initFloor
	e.doorTimer = time.Time{}
	e.hallRequests = make([][2]types.ButtonStatus, numFloors)
	e.cabRequests = make([]types.ButtonStatus, numFloors)

	e.IsMaster = false

	e.System.InitSystem(id, "192.168.0.1", 4)

	// e.elevatorState = types.ES_Uninitialized

	//Temp init door timer
	e.lostComsTimer = time.Time{}
	e.ackCounterLostComs = 0

	e.IpRegistery = make(map[string]string)

	// TODO temp
	e.SendToProtocol = make(chan message.ElevatorMessage, 10)

	e.hardwareEventsCh = make(chan HardwareEvent, 20)

	e.IsOnline = false
	// e.StatusChan = statusChan
	// e.TaskChan = taskChan

	// e.MsgRecieveCh = make(chan session.ElevatorPacket, 10) // Match the expected type

	// e.StatusChan <-utilities.StatusMsg{e.id, e.currentFloor, e.nextTarget}
	// e.MsgRecieveCh = make(chan message.ElevatorMessage, 10) // TODO Should have this in the code

	e.clearAllLamps(elevio.BT_HallUp, elevio.BT_HallDown, elevio.BT_Cab)
}

func (e *Elevator) RunElevatorProgram() {
	fmt.Println("RUNNING ELEVATOR PROGRAM")
	go e.RunHardwareEventLoop()
	go e.RunDoorStateMachine()
	go e.RunElevatorStateMachine()
	e.StartHardwareEventsListeners()
	// time.Sleep(10 * time.Second)

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
	id: %s
	in between floors: %t
	current floor: %d
	target: %d, %s
	init floor: %d
	last moving dir: %s
	door state: %s
	elevator state: %s
`,
		e.Id, e.inBetweenFloors, e.currentFloor, e.System.Elevators[e.Id].Target.Floor, e.System.Elevators[e.Id].Target.Button, e.initFloor, e.System.Elevators[e.Id].Direction, e.doorState, e.System.Elevators[e.Id].State)

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
