package transmissions

import (
	"fmt"
	"net"
	"time"
)

func SendMessage(remoteAddr string, message string) {
	remoteUDPAddr, err := net.ResolveUDPAddr(network, remoteAddr)
	if err != nil {
		fmt.Println(err)
	}

	conn, err := net.DialUDP(network, nil, remoteUDPAddr)
	if err != nil {
		fmt.Println(err)
	}
	defer conn.Close()

	msg := []byte(message)

	for k := 0; k < 5; k++ {
		_, err = conn.Write(msg)
		if err != nil {
			fmt.Println(err)
		}
		time.Sleep(2 * time.Second)
	}

	// fmt.Println("Sending successful")
}

func Broadcast() {}
