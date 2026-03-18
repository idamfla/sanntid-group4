package server

import (
	"elevator_program/udp"
	"elevator_program/udp/packet"
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

//	func (srv *Server) QueueElevatorTask(msgType message.MessageType, pkt packet.Packet, elevDone chan<- struct{}, taskReady <-chan struct{}) {
//		pkt.Payload = message.Message{
//			ID:      srv.ID,
//			Adder:   srv.recvConn.LocalAddr().String(),
//			MsgType: msgType,
//		}
func (srv *Server) QueueElevatorTask(pkt packet.Packet, elevDone chan<- struct{}, taskReady <-chan struct{}) {
	srv.elevatorTaskQueue <- ElevatorTask{
		ElevPacket: session.ElevatorPacket{
			Packet: pkt,
			Done:   elevDone,
		},
		Ready: taskReady,
	}
}

func (srv *Server) sendToElevator(elevTask ElevatorTask) {
	srv.elevator <- session.ElevatorPacket{
		Packet: elevTask.ElevPacket.Packet,
		Done:   elevTask.ElevPacket.Done,
	}
}

func (srv *Server) QueueElevatorSnapshot() {
	// TODO how to notify ...
}
