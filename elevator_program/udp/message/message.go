package message

type Message struct {
	Content string `json:"content,omitempty"`
}

// package message

// import "elevator_program/elevio"

// const (
// 	MSG_T_StatusReport MessageType = iota

// 	MSG_T_TaskCreate  // a new task is created/published
// 	MSG_T_ButtonPress // Slave notices a new button press
// 	// MSG_T_TaskAssign   // a task is assigned to you
// 	// MSG_T_TaskDelegate // a task is assigned to another person
// 	MSG_T_TaskUpdate   // task changed, Don't think we need it
// 	MSG_T_TaskComplete // task was completed
// 	MSG_T_TaskRequest  // someone requests a new task
// 	MSG_T_LostComs     // A routine to check if you have lost communication
// 	MSG_T_ElevatorLost // An elevator has lost coms, you need to send your connection to master status
// 	MSG_T_NewToChannel // Send the latest information
// )

// // Hopefully a better struct
// type Message struct {
// 	MsgType types.MessageType

// 	Id string
// 	Ip string

// 	// Elevator state reporting
// 	// Status *types.ElevatorsStatus // TODO Don't think we need this one, only need to use Elevator map
// 	ActivePeers int
// 	// Task / button updates
// 	Task      elevio.ButtonEvent // TODO do we want it as a pointer? Gives us the option to not send Task on every message
// 	BtnStatus types.ButtonStatus

// 	// System synchronization
// 	HallRequests [][2]types.ButtonStatus
// 	Elevators    map[string]types.ElevatorsStatus
// }
