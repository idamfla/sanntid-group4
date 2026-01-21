package elevator

func (s ElevatorState) String() string {
	switch s {
	// case Idle:
	// 		return "idle"
	case ES_Uninitialized:
		return "uninitialized"
	case ES_Idle:
			return "idle"
	case ES_Moving:
			return "moving"
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