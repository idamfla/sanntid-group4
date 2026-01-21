package task_manager

type TaskManager struct {
	UpButtons,DownButtons []int
	Elevators []*Elevator
	StatusChan chan StatusMsg
	TaskQueue chan TaskChan
	ElevatorDistanceToTarget [int]int // map of what task the elevators currently have assigned
}

func InitTaskManager() {
	
}