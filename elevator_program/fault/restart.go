package fault

import (
	"os"
	"os/exec"
)

func RestartSelf() {
	exe, err := os.Executable()
	if err != nil {
		os.Exit(1)
	}

	cmd := exec.Command(exe, os.Args[1:]...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	_ = cmd.Start()
	os.Exit(0)
}

