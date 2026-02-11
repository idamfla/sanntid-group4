package transmissions

import (
	"fmt"
	"net"
)

func Receiver() {
	// conn, err := net.Dial(network, address) //connects to the network
	conn, err := net.ListenPacket(network, ":30000")
	if err != nil {
		panic(err)
	}
	defer conn.Close() // when received message, close the connection right before returning function

	localAddr := conn.LocalAddr().(*net.UDPAddr) // returning the adress that the socket is bounded too (the local adress)

	fmt.Println("local address:", localAddr.String())

	buf := make([]byte, 1024)

	for {
		n, addr, err := conn.ReadFrom(buf) // reading data received from the buffer and storing the address where the data is from
		if err != nil {
			fmt.Println("error:", err)
			continue
		}

		// filter out own messages
		updAddr := addr.(*net.UDPAddr) // cast addr from net.Addr to *net.UDPAddr
		if updAddr.IP.Equal(localAddr.IP) && updAddr.Port == localAddr.Port {
			continue
		}

		fmt.Printf("%s: %s\n", updAddr.String(), string(buf[:n]))
	}
}
