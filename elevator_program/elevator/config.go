package elevator

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"elevator_program/elevio"
)


// Holds all runtime configuration values
type Config struct {
	ID        int
	IP        string
	Port      string
	Floors    int
	InitFloor int

	N        int
	BasePort int

	Addrs string
}

// Parses terminal arguments into a Config struct
func ParseFlags() Config {
	var cfg Config
	flag.IntVar(&cfg.ID, "id", 0, "elevator id (0 = launcher)")
	flag.StringVar(&cfg.IP, "ip", "localhost", "ip/host to elevio")
	flag.StringVar(&cfg.Port, "port", "15657", "elevio port (used when --id != 0)")
	flag.IntVar(&cfg.Floors, "floors", 4, "number of floors")
	flag.IntVar(&cfg.InitFloor, "initfloor", 3, "initial floor (0-index)")

	flag.IntVar(&cfg.N, "n", 3, "number of elevators to spawn (launcher only)")
	flag.IntVar(&cfg.BasePort, "baseport", 15657, "base port (launcher only)")

    flag.StringVar(&cfg.Addrs, "addrs", "", "comma-separated elevator addrs: ip:port,ip:port,... (launcher only)")

	flag.Parse()
	return cfg
}



// Starts a new process running one elevator instance
func startChildProcess(id int, ip, port string, floors, initFloor int) {
	cmd := exec.Command("go", "run", ".",
		"--id", strconv.Itoa(id),
		"--ip", ip,
		"--port", port,
		"--floors", strconv.Itoa(floors),
		"--initfloor", strconv.Itoa(initFloor),
	)

    // Forward child output to current terminal
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

    // Start process asynchronously
	if err := cmd.Start(); err != nil {
		fmt.Println("Cannot start the elevator", id, ":", err)
	}
}

// Launcher: spawns multiple elevator processes
func SpawnElevators(cfg Config) {
    // Mode A: explicit addresses (physical elevators)
	if cfg.Addrs != "" {
		addrs := strings.Split(cfg.Addrs, ",")
		for i, addr := range addrs {
			id := i + 1

			host, port, ok := strings.Cut(addr, ":")
			if !ok {
				fmt.Println("Bad addr (expected ip:port):", addr)
				continue
			}

			startChildProcess(id, host, port, cfg.Floors, cfg.InitFloor)
			fmt.Println("Started elevator", id, "on", addr)
		}
		return
}

	// Mode B: baseport + N (simulator on one machine)
	for i := 1; i <= cfg.N; i++ {

		// Assign consecutive ports
		port := strconv.Itoa(cfg.BasePort + (i - 1))

		startChildProcess(i, cfg.IP, port, cfg.Floors, cfg.InitFloor)

		fmt.Println("Started elevator", i, "on port", port)
	}
}


// Runs a single elevator instance
func RunOneElevator(cfg Config) {
	addr := fmt.Sprintf("%s:%s", cfg.IP, cfg.Port)
	elevio.Init(addr, cfg.Floors)

	var e Elevator
	e.InitElevator(cfg.ID, cfg.Floors, cfg.InitFloor)
	e.RunElevatorProgram()
}