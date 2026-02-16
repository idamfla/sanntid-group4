import (
	"context"
	"fmt"
	"net"
	"time"
	// "elevator_program/transmissions"
)

/*
Needs this somewhere
ctx, cancel := context.Withcancel(context.Background())
msgCh := make(chan string)
*/

// msg_type = broadcast, direct_msg, lost_coms, new_to_channel,
// msg = [type, port, btn_request[floor, dir], pos, dir]

func (e *Elevator) messageHandler(msgCh chan string) {
	for {
		select {
		case msg := <-msgCh:
			fmt.Println("Received: ", msg)
			types := msg[0]
			floor, dir = msg[2]

			switch types {
			case 0:
				// broadcast basic
				e.floorRequests[floor][dir] = true
			case 1:
				// direct msg
				e.nextTarget = {floor, dir}
			}

		case <-time.After(2 * time.Second):
			fmt.Println("Maybe lost comunication")
			//Broadcast that you have lost communication, figure out how to restart yourself or other
		}
		// Need to send ack 
	}
}
