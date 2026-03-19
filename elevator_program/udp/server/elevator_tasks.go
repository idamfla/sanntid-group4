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

	for task := range srv.elevatorTaskQueue {
		fmt.Println(srv.ID, "waiting for task to be ready")
		select {
		case <-task.Ready:
			fmt.Println(srv.ID, "task ready, sending to elevator")
			srv.sendToElevator(task)
		case <-time.After(udp.TASK_READY_TIMEOUT):
			fmt.Println(srv.ID, "task never became ready, skipping")
			return
		}
	}
	fmt.Println(srv.ID, "task loop stopped ...")
}

func (srv *Server) QueueElevatorTask(eMsg message.ElevatorMessage, elevDone chan<- struct{}, taskReady <-chan struct{}) {
	srv.elevatorTaskQueue <- ElevatorTask{
		ElevPacket: session.ElevatorPacket{
			EMsg: eMsg,
			Done: elevDone,
		},
		Ready: taskReady,
	}
}

func (srv *Server) sendToElevator(elevTask ElevatorTask) {
	srv.elevator <- session.ElevatorPacket{
		EMsg: elevTask.ElevPacket.EMsg,
		Done: elevTask.ElevPacket.Done,
	}
}
