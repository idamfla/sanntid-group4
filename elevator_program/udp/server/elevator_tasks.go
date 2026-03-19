package server

import (
	"elevator_program/message"
	"elevator_program/udp"
	"elevator_program/udp/session"
	"fmt"
	"time"
)

type ElevatorTask struct {
	ElevPacket session.ElevatorPacket
	Ready      <-chan struct{}
}

func (srv *Server) sendTaskLoop() {
	defer srv.wg.Done()

	for {
		select {
		case <-srv.stop:
			fmt.Println(srv.ID, "task loop stopped ...")
			return

		case task := <-srv.elevatorTaskQueue:
			fmt.Println(srv.ID, "waiting for task to be ready")

			select {
			case <-task.Ready:
				fmt.Println(srv.ID, "task ready, sending to elevator")
				srv.sendToElevator(task)

			case <-time.After(udp.TASK_READY_TIMEOUT):
				fmt.Println(srv.ID, "task never became ready, skipping")

			case <-srv.stop:
				fmt.Println(srv.ID, "task loop stopped during wait")
				return

			}
		}
	}
}

func (srv *Server) QueueElevatorTask(eMsg message.ElevatorMessage, elevDone chan<- struct{}, taskReady <-chan struct{}) {
	select {
	case srv.elevatorTaskQueue <- ElevatorTask{
		ElevPacket: session.ElevatorPacket{
			EMsg: eMsg,
			Done: elevDone,
		},
		Ready: taskReady,
	}:
	default:
		fmt.Println("Can't queue task, elevatorTaskQueue is full")
	}
}

func (srv *Server) sendToElevator(elevTask ElevatorTask) {
	select {
	case srv.elevator <- session.ElevatorPacket{
		EMsg: elevTask.ElevPacket.EMsg,
		Done: elevTask.ElevPacket.Done,
	}:
	default:
		fmt.Println("Can't send to elevator, toElevatorCh is full")
	}
}
