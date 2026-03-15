package system

import (
	"elevator_program/elevio"
	"elevator_program/message"
	"elevator_program/types"
)

// TODO Probably a bad name for the file

func (s *System) SetStatusReport(id int, elevator types.ElevatorsStatus) {
	s.Elevators[id] = elevator
}

func (s *System) SetRequestStatus(id int, status types.ButtonStatus, btnEvent elevio.ButtonEvent) {
	f := btnEvent.Floor // TODO Is it wierd that i define b and f?
	b := btnEvent.Button
	if b == elevio.BT_Cab {
		s.Elevators[id].CabRequests[f] = status
	} else {
		s.HallRequests[f][b] = status
	}
}

// func (s *System) UpdateRemoteCabBtn(id int, status types.ButtonStatus, floor int) {
// 	s.Elevators[id].CabRequests[floor] = status
// }

func (s *System) InitializeFromSystemState(msg message.Message) {
	// s.HallRequests = msg.HallRequests
	// s.Elevators = msg.Elevators

	s.HallRequests = make([][2]types.ButtonStatus, len(msg.HallRequests))
	copy(s.HallRequests, msg.HallRequests)

	s.Elevators = make(map[int]types.ElevatorsStatus)
	for id, e := range msg.Elevators {
		s.Elevators[id] = e
	}
}

func (s *System) RegisterAndSyncElevator(msg message.Message, ipRegistery map[string]int) (message.Message, int) {
	newMessage := message.Message{
		Ip: msg.Ip,
	}

	senderId, ok := ipRegistery[msg.Ip]
	if ok {
		newMessage.Id = msg.Id

	} else {
		senderId = findFreeID(s.Elevators)
		newElevator := types.ElevatorsStatus{
			Id: senderId,
			// Hope everything else is already configured
		}
		// TODO we also need to update IpRegistery
		s.Elevators[senderId] = newElevator
		newMessage.Id = newElevator.Id
	}
	newMessage.HallRequests = s.HallRequests
	newMessage.Elevators = s.Elevators // TODO send newMessage back
	return newMessage, senderId
}

// TODO Probably don't need only for testing
func (s System) CopySystem() System {
	// Create a new instance of System and copy the fields
	newCopy := s

	// Deep copy the map to ensure independence
	newCopy.Elevators = make(map[int]types.ElevatorsStatus)
	for id, elevator := range s.Elevators {
		newCopy.Elevators[id] = elevator
	}

	// Deep copy the hallRequests slice
	newCopy.HallRequests = make([][2]types.ButtonStatus, len(s.HallRequests))
	copy(newCopy.HallRequests, s.HallRequests)

	return newCopy
}

// TODO should I put this one in types?
func findFreeID(elevators map[int]types.ElevatorsStatus) int {
	id := 1
	for {
		if _, exists := elevators[id]; !exists {
			return id
		}
		id++
	}
}
