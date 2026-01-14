package elevator

type ElevatorStatus int

const (
	uninitialized ElevatorStatus = iota
	running
	doorOpen
	obstruction
	emergencyStop
)

func (s ElevatorStatus) String() string {
	switch s {
	// case Idle:
	// 		return "idle"
	case uninitialized:
		return "uninitialized"
	case running:
			return "Running"
	case doorOpen:
			return "door open"
	case obstruction:
			return "obstruction"
	case emergencyStop:
			return "emergency stop"
	default:
		return "unknown"
	}
}