// package elevtest

// import (
// 	"elevator_program/elevio"
// 	"sync"
// 	"time"
// )

// type Elevator struct {
// 	ID string

// 	// Hardware
// 	Motor   *Motor
// 	Door    *Door
// 	Buttons *ButtonPanel

// 	// // Logic
// 	// Orders *OrderManager

// 	// State (OWNED HERE)
// 	Status ElevatorStatus
// 	State  ElevatorState

// 	// Network
// 	Server *Server

// 	// View of other elevators (READ-ONLY COPY)
// 	ElevatorMap map[string]ElevatorView

// 	// Channels
// 	buttonCh chan elevio.ButtonEvent

// 	mu sync.Mutex
// }

// // what you know about yourself
// type ElevatorStatus struct {
// 	ID            string
// 	Floor         int
// 	Direction     Motor
// 	Target        int
// 	BetweenFloors bool
// }

// type Motor elevio.MotorDirection

// const (
// 	MTR_Up   Motor = Motor(elevio.MD_Up)
// 	MTR_Down Motor = Motor(elevio.MD_Down)
// )

// type ButtonPanel struct {
// 	Internal map[int]bool
// 	External map[int]map[elevio.ButtonType]bool
// }

// type Door struct {
// 	IsOpen       bool
// 	Obstructed   bool
// 	Timer        *time.Timer
// 	OpenDuration time.Duration

// 	mu sync.Mutex
// }

// // what you know about others
// type ElevatorView struct {
// 	ID        string
// 	Floor     int
// 	Direction Direction
// 	Target    int
// 	State     ElevatorState
// }

// // --- Msg ---

// type ElevatorStateMsg struct {
// 	ID        string
// 	Floor     int
// 	Direction Direction
// 	Target    int
// 	State     ElevatorState
// }

// // --- SERVER ---
// type PeerInfo struct {
// 	ID      string
// 	Address string

// 	Status PeerStatus

// 	LastSeen time.Time
// 	Alive    bool
// }

// type PeerStatus struct {
// 	Floor     int
// 	Direction Direction
// 	Target    int
// 	State     ElevatorState
// }

// /*
// Elevator
//  ├── Status (OWNED)
//  ├── Orders
//  ├── Hardware (Motor, Door, Buttons)
//  ├── ElevatorMap (VIEW of others)
//  └── Server (network interface)

// Server
//  ├── UDP + Sessions
//  ├── PeerInfo map (FULL truth)
//  ├── Master logic
//  └── Converts → ElevatorView

// PeerInfo
//  ├── Address
//  ├── LastSeen / Alive
//  └── PeerStatus

// PeerStatus
//  ├── Floor
//  ├── Direction
//  ├── Target
//  └── State
// */
