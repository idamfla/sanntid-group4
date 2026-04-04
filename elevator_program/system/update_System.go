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

	// Preserve master-managed fields (Target, CabRequests) so that
	// stale StatusReports don't overwrite newer task assignments.
	// if existing, exists := s.Elevators[ip]; exists {
	// 	elevator.Target = existing.Target
	// 	elevator.CabRequests = existing.CabRequests
	// }
	s.Elevators[ip] = elevator
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
		if elevatorCopy.Target.Floor == f && elevatorCopy.Target.Button == b {
			elevatorCopy.Target = elevio.ButtonEvent{
				Button: elevio.BT_HallUp,
				Floor:  -1,
			}
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

	// Update buttonlamp
	for f, row := range s.HallRequests {
		for b, btnStatus := range row {
			if btnStatus != types.NotActive {
				button := elevio.ButtonType(b)
				elevio.SetButtonLamp(button, f, true)
			}
		}
	}

	for f, btnStatus := range s.Elevators[msg.Addr].CabRequests {
		if btnStatus != types.NotActive {
			elevio.SetButtonLamp(elevio.BT_Cab, f, true)
		}
	}
}

func (s *System) RegisterAndSyncElevator(
	eMsg message.ElevatorMessage,
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
		CabRequests:    make([]types.ButtonStatus, numFloors),
		IsAlive:        true,
		IsMotorWorking: true,
	}

	if _, ok := s.Elevators[eMsg.Addr]; ok {
		old := s.Elevators[eMsg.Addr].CabRequests

		for i := range numFloors {
			if old[i] != types.NotActive || eMsg.Elevators[eMsg.Addr].CabRequests[i] != types.NotActive { // TODO does the new one be able to turn something to running??
				newElevator.CabRequests[i] = types.Pending
			}
		}
	}
	// TODO need this one for syncing
	// else {
	// 	copy(newElevator.CabRequests, eMsg.Elevators[eMsg.Addr].CabRequests)
	// }

	hallCopy := make([][2]types.ButtonStatus, len(s.HallRequests))
	copy(hallCopy, s.HallRequests) // TODO temp for without syncing

	// TODO need this one for syncing
	// for f, row := range s.HallRequests {
	// 	for b, btnStatus := range row {
	// 		if btnStatus == types.Running {
	// 			hallCopy[f][b] = types.Running
	// 		} else if btnStatus == types.Pending || eMsg.HallRequests[f][b] != types.NotActive {
	// 			hallCopy[f][b] = types.Pending
	// 		}
	// 	}
	// }

	s.HallRequests = hallCopy

	s.Elevators[eMsg.Addr] = newElevator

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
		oldTarget := s.Elevators[ip].Target
		// Only revert old target to Pending if it is actually Running.
		// This prevents phantom Pending from stale target data.
		if oldTarget.Button == elevio.BT_Cab {
			if s.Elevators[ip].CabRequests[oldTarget.Floor] == types.Running {
				s.SetRequestStatus(ip, types.Pending, oldTarget)
			}
		} else {
			if s.HallRequests[oldTarget.Floor][oldTarget.Button] == types.Running {
				s.SetRequestStatus(ip, types.Pending, oldTarget)
			}
		}
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
