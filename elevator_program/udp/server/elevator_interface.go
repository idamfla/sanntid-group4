package server

import (
	"elevator_program/message"
	"elevator_program/udp"
	"fmt"
	"time"
)

type ElevatorInterface struct {
	Recv      chan ElevatorPacket
	TaskQueue chan ElevatorTask
}

type ElevatorTask struct {
	ElevPacket ElevatorPacket
	Ready      <-chan struct{}
}

// Session -> Elevator
type ElevatorPacket struct {
	EMsg message.ElevatorMessage
	Done chan<- struct{}
}

func NewElevatorInterface(elevRecv chan ElevatorPacket) *ElevatorInterface {
	return &ElevatorInterface{
		Recv:      elevRecv,
		TaskQueue: make(chan ElevatorTask, CHANNEL_BUF),
	}
}

func (srv *Server) QueueElevatorTask(eMsg message.ElevatorMessage, elevDone chan<- struct{}, taskReady <-chan struct{}) {
	select {
	case srv.taskQueueCh() <- ElevatorTask{
		ElevPacket: ElevatorPacket{
			EMsg: eMsg,
			Done: elevDone,
		},
		Ready: taskReady,
	}:
	default:
		fmt.Println("Can't queue task, elevatorTaskQueue is full")
	}
}

func (srv *Server) taskQueueCh() chan ElevatorTask    { return srv.elevator.TaskQueue }
func (srv *Server) elevRecvCh() chan<- ElevatorPacket { return srv.elevator.Recv }

func (srv *Server) sendTaskLoop() {
	defer srv.WgDone()

	for {
		select {
		case <-srv.stopCh():
			fmt.Println(srv.GetAlias(), "task loop stopped ...")
			return

		case task := <-srv.taskQueueCh():
			fmt.Println(srv.GetAlias(), "waiting for task to be ready")

			select {
			case <-task.Ready:
				fmt.Println(srv.GetAlias(), "task ready, sending to elevator")
				srv.sendToElevator(task)

			case <-time.After(udp.TASK_READY_TIMEOUT):
				fmt.Println(srv.GetAlias(), "task never became ready, skipping")

			case <-srv.stopCh():
				fmt.Println(srv.GetAlias(), "task loop stopped during wait")
				return

			}
		}
	}
}

func (srv *Server) sendToElevator(elevTask ElevatorTask) {
	select {
	case <-srv.stopCh():
		fmt.Println("Server stopping, abort sending to elevator")
	case srv.elevRecvCh() <- elevTask.ElevPacket:
	default:
		fmt.Println("Can't send to elevator, toElevatorCh is full")
	}
}
