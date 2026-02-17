package elevator

import (
	"elevator_program/elevio"
	"fmt"
	"strconv"
	"time"
)

type MsgType int

const (
	MSG_T_Broadcast    MsgType = 0
	MSG_T_DirectMsg    MsgType = 1
	MSG_T_LostComs     MsgType = 2
	MSG_T_NewToChannel MsgType = 3
)

type Message struct {
	Id         string             // Master or not
	MsgType    int                // Type of msg
	Position   int                // Elevators current position
	NextTarget elevio.ButtonEvent // Elevators current target (floor, btnType) or change current target
	BtnMap     [][4]int           // Send cab calls to the other elevators, and if lost comunication the rest of the map too
}

// Temp need to create a map with IP and ID
var ID = make(map[string]int)

// TEMP map over cab calls
var cabs = make(map[int][4]int)

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

			id_num, err := strconv.Atoi(msg.Id) // Should do this somewhere where it does not happen when sending ip
			if err != nil {
				fmt.Println("Error when converting to int")
				return
			}

			switch msg.MsgType {
			case 0:
				// broadcast basic
				// If there is a new request, need to add it to button map, turn on lights, send ack
			case 1:
				// direct msg
				if id_num == 0 {
					// Need to be sure that the overwritten target dose'nt get lost
					e.nextTarget = msg.NextTarget
					// Need to send ack to master
				} else {
					// Master need to add request to que and inform other elevators
				}
			case 2:
				// lost coms
				// Meesage your communication status with the master
			case 3:
				// New to chanel
				// Check if the IP has been here before, if this is the case, send back the id and button map
				// and all the cab calls
			}

		case <-time.After(2 * time.Second):
			fmt.Println("Maybe lost comunication")
			//Broadcast that you have lost communication, figure out how to restart yourself or other
		}
	}
}
