package utilities

type Motion int

const (
	MotionStop Motion = iota
	MotionMoveUp
	MotionMoveDown
)

func (m Motion) String() string {
	switch m {
	case MotionStop:
		return "stop"
	case MotionMoveUp:
		return "move up"
	case MotionMoveDown:
		return "move down"
	default:
		return "unknown"
	}
}