package nettwork

import (
	"context"
	"fmt"
	"net"
	"time"
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

// Trancieves message on port with specific type
func trancive(msg string, port string, typeOfMessage string, network string) {
	addr := typeOfMessage + ":" + port
	remoteUDPAddr, err := net.ResolveUDPAddr(network, addr)
	if err != nil {
		fmt.Println(err)
	}

	conn, err := net.DialUDP(network, nil, remoteUDPAddr)
	if err != nil {
		fmt.Println(err)
	}
	defer conn.Close()

	message := []byte(msg)
	_, err = conn.Write(message)
	if err != nil {
		fmt.Println(err)
	}
}

// receives message from port, sends it through a chanel, checks if it lost comunication
func receiver(ctx context.Context, port string, network string, msgCh chan<- string) error {
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

			msg := string(buf[:n])
			msgCh <- msg // send message to main goroutine or handler
		}
	}
}

/*
Needs this somewhere
ctx, cancel := context.Withcancel(context.Background())
msgCh := make(chan string)
*/

func messageHandler(msgCh chan string) {
	for {
		select {
		case msg := <-msgCh:
			fmt.Println("Received: ", msg)
		case <-time.After(2 * time.Second):
			fmt.Println("Maybe lost comunication")
			//Broadcast that you have lost communication, figure out how to restart yourself or other
		}
	}
}
