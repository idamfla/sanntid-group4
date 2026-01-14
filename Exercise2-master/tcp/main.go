package main

import (
	"fmt"
	"net"
	"time"
)

const serverIP = "10.100.23.11"
const serverPort = "33546" // \0
// const port = "34933" // fixed size, 1024

const localIP = "10.100.23.15"
const localPort = "20005"

const network = "tcp"

func TCPListen(conn net.Conn) {
	buf := make([]byte, 2048)

	for {
		n, err := conn.Read(buf)

		if err != nil {
			fmt.Println("Server closed connection:", err)
			return
		}

		fmt.Println(string(buf[:n]))
	}
}

func TCPWrite(conn net.Conn, msg string) error {
	if conn == nil {
		fmt.Println("Error: nil connection")
		return nil
	}

	data := []byte(msg + "\x00")
	total := 0

	for total < len(data) {
		n, err := conn.Write(data[total:])
		if err != nil {
			fmt.Println("Error with writing:", err)
			return err
		}
		total += n
	}
	time.Sleep(2 * time.Second)
	return nil
}

func main() {
	conn, err := net.Dial(network, serverIP+":"+serverPort)

	if err != nil {
		panic(err)
	}

	defer conn.Close()

	fmt.Println("Server listening on port", serverPort)

	go TCPListen(conn)

	// go TCPWrite(conn, "Connect to: "+localIPAddr+":"+port+"\x00")
	// go TCPWrite(conn, "Connect to: "+localIPAddr+":"+localPort+"\x00")

	localPort := conn.LocalAddr().(*net.TCPAddr).Port
	// msg := fmt.Sprintf("IPv4 is: %s:%d", localIP, localPort)
	// err = TCPWrite(conn, msg)
	messages := []string{
		fmt.Sprintf("Connect to: %s:%d", localIP, localPort),
		"Second message",
		"Third message",
	}

	for _, msg := range messages {
		err = TCPWrite(conn, msg)
		if err != nil {
			fmt.Println("Error sending message:", err)
			continue
		}
	}

	select {}
}
