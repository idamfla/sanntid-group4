package main

import (
	"upd/transmissions"
)

const ipAddr = "10.100.23.15"
const localPort = "30000"
const remotePort = "20005"

func main() {
	addr := ipAddr + ":" + localPort

	go transmissions.Receiver()
	go transmissions.SendMessage(addr, "Hello from group 4")
	select {}
}
