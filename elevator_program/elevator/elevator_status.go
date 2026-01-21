package elevator

type ElevatorState int

const (
	ES_Uninitialized ElevatorState = iota
	ES_Running
	ES_DoorOpen
	ES_Obstruction
	ES_EmergencyStop
)

func (s ElevatorState) String() string {
	switch s {
	// case Idle:
	// 		return "idle"
	case ES_Uninitialized:
		return "uninitialized"
	case ES_Running:
			return "running"
	case ES_DoorOpen:
			return "door open"
	case ES_Obstruction:
			return "obstruction"
	case ES_EmergencyStop:
			return "emergency stop"
	default:
		return "unknown"
	}
}