package system

import (
	"elevator_program/elevio"
	"elevator_program/types"
	"sync"
)

// Uses to add a new elevator to our system
type System struct {
	HallRequests [][2]types.ButtonStatus
	Elevators    map[string]types.ElevatorsStatus
	Mutex        sync.RWMutex // Add a mutex to protect shared data
}

func (s *System) InitSystem(id string, ip string, numFloors int) {
	if s.Elevators == nil {
		s.Elevators = make(map[string]types.ElevatorsStatus)
	}
	if s.HallRequests == nil {
		s.HallRequests = make([][2]types.ButtonStatus, numFloors)
	}

	s.Elevators[id] = types.ElevatorsStatus{
		Id:           id,
		Ip:           ip,
		CurrentFloor: -1,
		CabRequests:  make([]types.ButtonStatus, numFloors),
		Target: elevio.ButtonEvent{
			Floor:  -1,
			Button: elevio.BT_HallUp,
		},
		State: types.ES_Uninitialized,
	}
}

// Snapshot returns a deep copy of the system state. Caller MUST hold s.Mutex (RLock or Lock).
func (s *System) Snapshot() (hall [][2]types.ButtonStatus, elevs map[string]types.ElevatorsStatus) {
	hall = make([][2]types.ButtonStatus, len(s.HallRequests))
	copy(hall, s.HallRequests)

	elevs = make(map[string]types.ElevatorsStatus)
	for id, e := range s.Elevators {
		elevs[id] = e
	}
	return
}
