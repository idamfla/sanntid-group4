package elevator

func (e *Elevator) EventLoop() {
	for ev := range e.eventsCh {
		// optionally filter invalid events
		if ev.Type == EV_FloorSensor && ev.Floor == -1 {
			continue
		}
		// forward to state machine
		e.stateMachineCh <- ev
	}

	// region old
	// for {
	// 	var ev ElevatorEvent
	// 	select {
	// 	case ev = <-e.eventsCh:
	// 		switch ev.Type {
	// 		case EV_FloorSensor:
	// 			if ev.Floor != -1 {
	// 				e.currentFloor = ev.Floor
	// 			}
	// 		case EV_ButtonPress:
	// 			e.floorRequests[ev.Floor][ev.Button] = true
	// 			elevio.SetButtonLamp(ev.Button, ev.Floor, true)
	// 			// TODO send press to master
	// 		case EV_Obstruction:
	// 			if ev.Obstruction {
	// 				e.state = ES_Obstruction
	// 			}
	// 		case EV_EmergencyStop:
	// 			if ev.EmergencyStop {
	// 				elevio.SetStopLamp(true)
	// 				e.state = ES_EmergencyStop
	// 				continue
	// 			} else {
	// 				elevio.SetStopLamp(false)
	// 				e.state = ES_Idle
	// 			}
	// 		case EV_TaskAssigned:
	// 			continue
	// 		case EV_TaskCompleted:
	// 			continue
	// 		}
	// 	}
	// }
	// endregion
}