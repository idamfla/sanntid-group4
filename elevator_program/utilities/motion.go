package utilities

type Motion int

const (
	Stop Motion = iota
	MoveUp
	MoveDown
)

func (m Motion) String() string {
	switch m {
	case Stop:
		return "stop"
	case MoveUp:
		return "move up"
	case MoveDown:
		return "move down"
	default:
		return "unknown"
	}
}