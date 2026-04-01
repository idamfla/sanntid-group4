package system

import (
	"elevator_program/elevio"
	"elevator_program/message"
	"elevator_program/types"
	"fmt"
)

func (s *System) SetStatusReport(ip string, elevator types.ElevatorsStatus) {
	s.Mutex.Lock()
	defer s.Mutex.Unlock()
	s.Elevators[ip] = elevator
	fmt.Println("Help meeee \n\n\n\n\n", s)
}

func (s *System) SetRequestStatus(ip string, status types.ButtonStatus, btnEvent elevio.ButtonEvent) {
	f := btnEvent.Floor
	b := btnEvent.Button
	if b == elevio.BT_Cab {
		elevatorCopy := s.Elevators[ip]
		elevatorCopy.CabRequests[f] = status
		s.Elevators[ip] = elevatorCopy
	} else {
		s.HallRequests[f][b] = status
	}

	if status == types.NotActive {
		elevatorCopy := s.Elevators[ip]
		elevatorCopy.Target = elevio.ButtonEvent{
			Button: elevio.BT_HallUp,
			Floor:  -1,
		}
		s.Elevators[ip] = elevatorCopy
	}
}

func (s *System) InitializeFromSystemState(msg message.ElevatorMessage) {
	s.Mutex.Lock()
	defer s.Mutex.Unlock()

	s.HallRequests = make([][2]types.ButtonStatus, len(msg.HallRequests))
	copy(s.HallRequests, msg.HallRequests)

	s.Elevators = make(map[string]types.ElevatorsStatus)
	for ip, e := range msg.Elevators {
		if ip == e.Ip {

		}
		s.Elevators[ip] = e
	}
}

func (s *System) RegisterAndSyncElevator(
	eMsg message.ElevatorMessage,
	// ipRegistery map[string]string,
	numFloors int,
) message.ElevatorMessage {

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

	if _, ok := s.Elevators[eMsg.Addr]; ok {
		old := s.Elevators[eMsg.Addr].CabRequests

		newElevator.CabRequests = make([]types.ButtonStatus, len(old))
		copy(newElevator.CabRequests, old)
	}

	s.Elevators[eMsg.Addr] = newElevator

	hallCopy := make([][2]types.ButtonStatus, len(s.HallRequests))
	copy(hallCopy, s.HallRequests)

	elevCopy := make(map[string]types.ElevatorsStatus)
	for ip, e := range s.Elevators {
		elevCopy[ip] = e
	}

	newMessage.HallRequests = hallCopy
	newMessage.Elevators = elevCopy

	fmt.Println("Trying to sync, \n\n\n", newMessage)

	return newMessage
}

func (s *System) IsRequestInSystem(ip string, task elevio.ButtonEvent) bool {
	s.Mutex.RLock()
	defer s.Mutex.RUnlock()
	f := task.Floor
	b := task.Button
	if b == elevio.BT_Cab {
		return s.Elevators[ip].CabRequests[f] != types.NotActive
	} else {
		return s.HallRequests[f][b] != types.NotActive
	}
}

func (s *System) SetRequestAsTarget(ip string, task elevio.ButtonEvent) {
	s.Mutex.Lock()
	defer s.Mutex.Unlock()

	if s.Elevators[ip].Target.Floor != -1 {
		s.SetRequestStatus(ip, types.Pending, s.Elevators[ip].Target)
	}

	s.SetRequestStatus(ip, types.Running, task)

	elevatorCopy := s.Elevators[ip]
	elevatorCopy.Target = task

	if task.Floor > s.Elevators[ip].CurrentFloor {
		elevatorCopy.Direction = elevio.MD_Up
	} else if task.Floor < s.Elevators[ip].CurrentFloor {
		elevatorCopy.Direction = elevio.MD_Down
	}
	s.Elevators[ip] = elevatorCopy
}
