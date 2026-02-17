package network

import (
	"elevator_program/elevator"
	"strconv"
	"time"
)

func selectProtocol()

// Sending all information the new elevator needs to work
func connectElevToSystem(port string, Id_self int, msg elevator.Message, cabs map[int][4]int, floorRequests [][3]int, ack chan elevator.Message) {
	msg.Id = strconv.Itoa(Id_self) // Need to convert to string

	id_num, exists := elevator.ID[msg.Id]
	if exists {
		msg.BtnMap[0] = cabs[id_num] // Had to put in as a parameter for it to work, thought of just collecting from elevator package
	}

	for i := 1; i < len(floorRequests); i++ {
		msg.BtnMap[1][i] = floorRequests[i][1] // Btns pushed upwards
		msg.BtnMap[2][i] = floorRequests[i][2] // Btns pushed downwards
	}

	for i := 0; i < 3; i++ { // retry 3 times
		trancive(msg, port, "unicom", "udp4")

		select {
		case <-ack:
			println("Ack received")

		case <-time.After(200 * time.Millisecond):
			println("Retrying...")
		}
	}

}
