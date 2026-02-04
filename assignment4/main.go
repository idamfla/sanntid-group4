package main

import (
	"assignment4/process_pair"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"time"
)

func startBackup(port int) {
	cmd := exec.Command("./main", "backup", strconv.Itoa(port))
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Start(); err != nil {
		panic(err)
	}

	fmt.Println("Started backup on port", port)
}

func main() {
	if len(os.Args) < 2 {
		// Initial startup
		backupPort := 9999
		startBackup(backupPort)

		time.Sleep(200 * time.Millisecond)

		process_pair.Primary("127.0.0.1:"+strconv.Itoa(backupPort), 1)
		return
	}

	mode := os.Args[1]

	switch mode {
	case "backup":
		port, _ := strconv.Atoi(os.Args[2])
		addr := "127.0.0.1:" + strconv.Itoa(port)
		process_pair.Backup(addr)
	}
}
