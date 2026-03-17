package system

import (
	"elevator_program/elevio"
	"elevator_program/message"
	"elevator_program/types"
	"fmt"
)

// TODO Probably a bad name for the file

func (s *System) SetStatusReport(id string, elevator types.ElevatorsStatus) {
	fmt.Println("System befor error: ", s)
	s.Elevators[id] = elevator
}

func (s *System) SetRequestStatus(id string, status types.ButtonStatus, btnEvent elevio.ButtonEvent) {
	f := btnEvent.Floor // TODO Is it wierd that i define b and f?
	b := btnEvent.Button
	if b == elevio.BT_Cab {
		fmt.Println("System before error: ", s)
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

	s.Elevators = make(map[string]types.ElevatorsStatus)
	for id, e := range msg.Elevators {
		if id == e.Id {

		}
		s.Elevators[id] = e
	}
}

func (s *System) RegisterAndSyncElevator(msg message.Message, ipRegistery map[string]string) (message.Message, string) {
	newMessage := message.Message{
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

	_, ok := ipRegistery[msg.Ip]
	if ok {
		for f, btnStatus := range s.Elevators[msg.Id].CabRequests {
			if btnStatus != types.NotActive || msg.Elevators[msg.Id].CabRequests[f] != types.NotActive {
				newElevator.CabRequests[f] = types.Pending
			}
		}
	} else {
		for f, btnStatus := range msg.Elevators[msg.Id].CabRequests {
			if btnStatus != types.NotActive {
				newElevator.CabRequests[f] = types.Pending
			}
		}
		s.Elevators[msg.Id] = newElevator
	}
	newMessage.HallRequests = s.HallRequests
	newMessage.Elevators = s.Elevators
	return newMessage, msg.Id
}

// TODO Probably don't need only for testing
func (s System) CopySystem() System {
	// Create a new instance of System and copy the fields
	newCopy := s

	// Deep copy the map to ensure independence
	newCopy.Elevators = make(map[string]types.ElevatorsStatus)
	for id, elevator := range s.Elevators {
		newCopy.Elevators[id] = elevator
	}

	// Deep copy the hallRequests slice
	newCopy.HallRequests = make([][2]types.ButtonStatus, len(s.HallRequests))
	copy(newCopy.HallRequests, s.HallRequests)

	return newCopy
}
