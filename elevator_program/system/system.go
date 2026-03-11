package system

import (
	"elevator_program/types"
)

// Uses to add a new elevator to our system
type System struct {
	HallRequests [][2]types.ButtonStatus
	Elevators    map[int]types.ElevatorsStatus
	// mutex        sync.Mutex // Add a mutex to protect shared data
}

// func (s *System) InitSystem(numFloors int) {
// 	s.HallRequests = make([][2]types.ButtonStatus, numFloors)
// 	s.Elevators = make(map[int]types.ElevatorsStatus)
// 	s.Elevators[id] = types.ElevatorsStatus{
// 		CabRequests: make([]types.ButtonStatus, numFloors),
// 		Id:          id,
// 	}
// }
