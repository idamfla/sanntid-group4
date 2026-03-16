package config

import (
	"elevator_program/elevator"
	"elevator_program/elevio"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
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

// ---------- Helpers for spawning in separate terminal windows ----------

func commandExists(name string) bool {
	_, err := exec.LookPath(name)
	return err == nil
}

func buildElevatorCmd(id int, ip, port string, floors, initFloor int) string {
	return fmt.Sprintf(
		`go run . --id %d --ip %s --port %s --floors %d --initfloor %d`,
		id, ip, port, floors, initFloor,
	)
}

// Starts a new OS process running one elevator instance,
// in its OWN terminal window if possible.
func startChildProcess(id int, ip, port string, floors, initFloor int) {
	wd, err := os.Getwd()
	if err != nil {
		fmt.Println("Cannot get working directory:", err)
		return
	}

	cmdStr := fmt.Sprintf(
		`go run . --id %d --ip %s --port %s --floors %d --initfloor %d`,
		id, ip, port, floors, initFloor,
	)

	// gnome-terminal: start in the SAME directory as launcher
	if commandExists("gnome-terminal") {
		cmd := exec.Command(
			"gnome-terminal",
			"--working-directory", wd,
			"--",
			"bash", "-c",
			cmdStr+`; echo; echo '--- exited (press enter) ---'; read`,
		)
		if err := cmd.Start(); err != nil {
			fmt.Println("Cannot start elevator", id, "(gnome-terminal):", err)
		}
		return
	}

	// xterm fallback
	if commandExists("xterm") {
		cmd := exec.Command(
			"xterm",
			"-T", fmt.Sprintf("Elevator %d", id),
			"-e", "bash", "-c",
			"cd "+strconv.Quote(wd)+"; "+cmdStr+`; echo; echo '--- exited (press enter) ---'; read`,
		)
		if err := cmd.Start(); err != nil {
			fmt.Println("Cannot start elevator", id, "(xterm):", err)
		}
		return
	}

	// Last resort: same terminal
	fmt.Println("No terminal emulator found. Running in same terminal.")
	cmd := exec.Command("go", "run", ".",
		"--id", strconv.Itoa(id),
		"--ip", ip,
		"--port", port,
		"--floors", strconv.Itoa(floors),
		"--initfloor", strconv.Itoa(initFloor),
	)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
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
		port := strconv.Itoa(cfg.BasePort + (i - 1))
		startChildProcess(i, cfg.IP, port, cfg.Floors, cfg.InitFloor)
		fmt.Println("Started elevator", i, "on", cfg.IP+":"+port)
	}
}

// Runs a single elevator instance
func RunOneElevator(cfg Config) {
	addr := fmt.Sprintf("%s:%s", cfg.IP, cfg.Port)
	elevio.Init(addr, cfg.Floors)

	var e elevator.Elevator
	e.InitElevator("Her skal det stå cfg.ID", cfg.Floors, cfg.InitFloor, "127.0.0.10", 9000) // TODO Probably need to change ip and port
	e.RunElevatorProgram()
}
