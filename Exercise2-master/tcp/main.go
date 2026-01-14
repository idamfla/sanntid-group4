package main

import (
	"fmt"
	"net"
)

const ipAddr = "10.100.23.11"

// const port = "34933" // fixed size, 1024
const port = "33546" // \0

const network = "tcp"

func main() {
	conn, err := net.Dial(network, ipAddr+":"+port)
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	fmt.Println("Connected to server")
}
