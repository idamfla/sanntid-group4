package utilities

import (
	"fmt"
)

type TaskMsg struct {
	Floor int
	Motion Motion
}

func (t TaskMsg) String() string {
	return fmt.Sprintf(`Task Message:
	floor: %d
	motion: %s`,
	t.Floor, t.Motion)
}