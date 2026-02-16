package elevator

import (
	"elevator_program/elevio"
	"fmt"
	"time"
	// "elevator_program/transmissions"
)

type MsgType int

const (
	MSG_T_Broadcast    MsgType = 0
	MSG_T_DirectMsg    MsgType = 1
	MSG_T_LostComs     MsgType = 2
	MSG_T_NewToChannel MsgType = 3
)

type Message struct {
	Id         int                // Master or not
	MsgType    int                // Type of msg
	Position   int                // Elevators current position
	NextTarget elevio.ButtonEvent // Elevators current target (floor, btnType) or change current target
	CabCalls   [4]int             // Send cab calls to the other elevators
}

// Temp need to create a map with IP and ID
var ID = make(map[string]int)

/*
Needs this somewhere
ctx, cancel := context.Withcancel(context.Background())
msgCh := make(chan string)
*/

func (e *Elevator) messageHandler(msgCh chan Message) {
	for {
		select {
		case msg := <-msgCh:
			fmt.Println("Received msg")
			switch msg.MsgType {
			case 0:
				// broadcast basic
				// If there is a new request, need to add it to button map, turn on lights, send ack
			case 1:
				// direct msg
				if msg.Id == 0 {
					// Need to be sure that the overwritten target dose'nt get lost
					e.nextTarget = msg.NextTarget
					// Need to send ack to master
				} else {
					// Master need to add request to que and inform other elevators
				}
			}

		case <-time.After(2 * time.Second):
			fmt.Println("Maybe lost comunication")
			//Broadcast that you have lost communication, figure out how to restart yourself or other
		}
	}
}
