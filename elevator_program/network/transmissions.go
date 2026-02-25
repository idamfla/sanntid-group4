package network

import (
	"context"
	"elevator_program/elevator"
	"encoding/json"
	"fmt"
	"net"
)

type ID int

var ports = map[ID]int{
	0: 30000, // Broadcast
}

type typeOfMessage string

var IP = map[typeOfMessage]string{
	"broadcast": "255.255.255.255",
	"unicom":    "127.0.0.1",
}

// Maybe change name of network is always ment as "udp4"
// Trancieves message on port with specific type
// Should maybe return bool
func Trancive(msg elevator.Message, port string, typeOfMessage string, network string) {
	addr := typeOfMessage + ":" + port
	remoteUDPAddr, err := net.ResolveUDPAddr(network, addr)
	if err != nil {
		fmt.Println(err)
		return
	}

	conn, err := net.DialUDP(network, nil, remoteUDPAddr)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer conn.Close()

	// Need to convert so it is possible to send struct through udp
	message, err := json.Marshal(msg)
	if err != nil {
		fmt.Println("Marshal error:", err)
		return
	}

	_, err = conn.Write(message)
	if err != nil {
		fmt.Println(err)
		return
	}
}

// receives message from port, sends it through a chanel, checks if it lost comunication
func Receiver(ctx context.Context, port string, network string, msgCh chan<- elevator.Message) error {
	conn, err := net.ListenPacket(network, ":"+port)
	if err != nil {
		panic(err)
	}
	defer conn.Close() // Will close when function exits

	localAddr := conn.LocalAddr().(*net.UDPAddr) // returning the adress that the socket is bounded too (the local adress)
	buf := make([]byte, 1024)

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()

		default:
			n, addr, err := conn.ReadFrom(buf)
			if err != nil {
				// other errors
				fmt.Println(err)
				continue
			}

			// Maybe delete
			// filter out own messages
			updAddr := addr.(*net.UDPAddr) // cast addr from net.Addr to *net.UDPAddr
			if updAddr.IP.Equal(localAddr.IP) && updAddr.Port == localAddr.Port {
				continue
			}

			var msg elevator.Message

			err = json.Unmarshal(buf[:n], &msg)
			if err != nil {
				fmt.Println("Unmarshal error:", err) // Convert back to struct
				continue
			}
			msgCh <- msg
		}
	}
}
