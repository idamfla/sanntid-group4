package elevator

import (
	"fmt"
	"time"
)

func (e *Elevator) ElevatorStateMachine() {
	ticker := time.NewTicker(50 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case ev := <-e.eventsCh:
			e.handleEvent(ev)
		case <-ticker.C:
			e.updateMotor()
		}
		fmt.Println(e) // DB
	}
}
