package elevator

type ElevatorState int

const (
	ElevStateUninitialized ElevatorState = iota
	ElevStateRunning
	ElevStateDoorOpen
	ElevStateObstruction
	ElevStateEmergencyStop
)

func (s ElevatorState) String() string {
	switch s {
	// case Idle:
	// 		return "idle"
	case ElevStateUninitialized:
		return "uninitialized"
	case ElevStateRunning:
			return "running"
	case ElevStateDoorOpen:
			return "door open"
	case ElevStateObstruction:
			return "obstruction"
	case ElevStateEmergencyStop:
			return "emergency stop"
	default:
		return "unknown"
	}
}