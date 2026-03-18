package system

import (
	"elevator_program/elevio"
	"elevator_program/message"
	"elevator_program/types"
)

// TODO Probably a bad name for the file

func (s *System) SetStatusReport(id string, elevator types.ElevatorsStatus) {
	s.Mutex.Lock()
	defer s.Mutex.Unlock()
	s.Elevators[id] = elevator
}

func (s *System) SetRequestStatus(id string, status types.ButtonStatus, btnEvent elevio.ButtonEvent) {
	f := btnEvent.Floor
	b := btnEvent.Button
	if b == elevio.BT_Cab {
		elevatorCopy := s.Elevators[id]
		elevatorCopy.CabRequests[f] = status
		s.Elevators[id] = elevatorCopy
	} else {
		s.HallRequests[f][b] = status
	}
}

func (s *System) InitializeFromSystemState(msg message.ElevatorMessage) {
	s.Mutex.Lock()
	defer s.Mutex.Unlock()

	s.HallRequests = make([][2]types.ButtonStatus, len(msg.HallRequests))
	copy(s.HallRequests, msg.HallRequests)

	s.Elevators = make(map[string]types.ElevatorsStatus)
	for id, e := range msg.Elevators {
		if id == e.Id {

		}
		s.Elevators[id] = e
	}
}

func (s *System) RegisterAndSyncElevator(msg message.ElevatorMessage, ipRegistery map[string]string) (message.ElevatorMessage, string) {
	newMessage := message.ElevatorMessage{
		MsgType: types.MSG_T_NewToChannel,
		Id:      msg.Id,
		Ip:      msg.Ip,
	}
	newElevator := types.ElevatorsStatus{
		Id:       msg.Id,
		Ip:       msg.Ip,
		IsMaster: false,
		Target: elevio.ButtonEvent{
			Floor:  -1,
			Button: elevio.BT_HallUp,
		},
		CabRequests: make([]types.ButtonStatus, len(msg.Elevators[msg.Id].CabRequests)),
	}

	s.Mutex.Lock()
	defer s.Mutex.Unlock()

	_, ok := ipRegistery[msg.Ip] // TODO Maybe remove the way to not loose new elevators requests if we don't use online
	if ok {
		newElevator.CabRequests = s.Elevators[msg.Id].CabRequests
		// for f, btnStatus := range s.Elevators[msg.Id].CabRequests {
		// 	if btnStatus != types.NotActive || msg.Elevators[msg.Id].CabRequests[f] != types.NotActive {
		// 		newElevator.CabRequests[f] = types.Pending
		// 	}
		// }
	}

	s.Elevators[msg.Id] = newElevator
	hall, elevs := s.Snapshot()
	newMessage.HallRequests = hall
	newMessage.Elevators = elevs
	return newMessage, msg.Id
}

// func (s *System) CopySystem() *System {
// 	// s.Mutex.RLock()
// 	// defer s.Mutex.RUnlock()

// 	// Create a new instance of System and copy the fields
// 	newCopy := &System{} // Create a pointer to a new System

// 	// Deep copy the map to ensure independence
// 	newCopy.Elevators = make(map[string]types.ElevatorsStatus)
// 	for id, elevator := range s.Elevators {
// 		newCopy.Elevators[id] = elevator
// 	}

// 	// Deep copy the hallRequests slice
// 	newCopy.HallRequests = make([][2]types.ButtonStatus, len(s.HallRequests))
// 	copy(newCopy.HallRequests, s.HallRequests)

// 	return newCopy
// }

func (s *System) IsRequestInSystem(id string, task elevio.ButtonEvent) bool {
	s.Mutex.RLock()
	defer s.Mutex.RUnlock()
	f := task.Floor
	b := task.Button
	if b == elevio.BT_Cab {
		return s.Elevators[id].CabRequests[f] != types.NotActive
	} else {
		return s.HallRequests[f][b] != types.NotActive
	}
}

func (s *System) SetRequestAsTarget(id string, task elevio.ButtonEvent) {
	// TODO I think it is wierd that I call system from here. The whole purpose of this was to seperate sytsem and elevator
	s.Mutex.Lock()
	defer s.Mutex.Unlock()

	if s.Elevators[id].Target.Floor != -1 {
		s.SetRequestStatus(id, types.Pending, s.Elevators[id].Target)
	}

	s.SetRequestStatus(id, types.Running, task)

	elevatorCopy := s.Elevators[id]
	elevatorCopy.Target = task

	if task.Floor > s.Elevators[id].CurrentFloor {
		elevatorCopy.Direction = elevio.MD_Up
	} else if task.Floor < s.Elevators[id].CurrentFloor {
		elevatorCopy.Direction = elevio.MD_Down
	}
	s.Elevators[id] = elevatorCopy
}
