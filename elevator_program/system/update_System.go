package system

import (
	"elevator_program/elevio"
	"elevator_program/message"
	"elevator_program/types"
	"fmt"
)

func (s *System) SetStatusReport(id string, elevator types.ElevatorsStatus) {
	s.Mutex.Lock()
	defer s.Mutex.Unlock()
	s.Elevators[id] = elevator
	fmt.Println("Help meeee \n\n\n\n\n", s)
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

	if status == types.NotActive {
		elevatorCopy := s.Elevators[id]
		elevatorCopy.Target = elevio.ButtonEvent{
			Button: elevio.BT_HallUp,
			Floor:  -1,
		}
		s.Elevators[id] = elevatorCopy
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

func (s *System) RegisterAndSyncElevator(
	eMsg message.ElevatorMessage,
	ipRegistery map[string]string,
	numFloors int,
) (message.ElevatorMessage, string) {

	s.Mutex.Lock()
	defer s.Mutex.Unlock()

	newMessage := message.ElevatorMessage{
		EMsgType: message.EMSG_T_NewToChannel,
		ID:       eMsg.ID,
		Addr:     eMsg.Addr,
	}

	newElevator := types.ElevatorsStatus{
		Id: eMsg.ID,
		Ip: eMsg.Addr,
		Target: elevio.ButtonEvent{
			Floor:  -1,
			Button: elevio.BT_HallUp,
		},
		CabRequests: make([]types.ButtonStatus, numFloors),
	}

	if _, ok := ipRegistery[eMsg.Addr]; ok {
		old := s.Elevators[eMsg.ID].CabRequests

		newElevator.CabRequests = make([]types.ButtonStatus, len(old))
		copy(newElevator.CabRequests, old)
	}

	s.Elevators[eMsg.ID] = newElevator

	hallCopy := make([][2]types.ButtonStatus, len(s.HallRequests))
	copy(hallCopy, s.HallRequests)

	elevCopy := make(map[string]types.ElevatorsStatus)
	for id, e := range s.Elevators {
		elevCopy[id] = e
	}

	newMessage.HallRequests = hallCopy
	newMessage.Elevators = elevCopy

	fmt.Println("Trying to sync, \n\n\n", newMessage)

	return newMessage, eMsg.ID
}

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
