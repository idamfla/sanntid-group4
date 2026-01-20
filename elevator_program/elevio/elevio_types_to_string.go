package elevio

func (md MotorDirection) String() string {
	switch md {
	case MD_Stop:
		return "motor stop"
	case MD_Up:
		return "motor up"
	case MD_Down:
		return "motor down"
	default:
		return "unknown"
	}
}

func (bt ButtonType) String() string {
	switch bt {
	case BT_Cab:
		return "button cab"
	case BT_HallUp:
		return "button up"
	case BT_HallDown:
		return "button down"
	default:
		return "unknown"
	}
}