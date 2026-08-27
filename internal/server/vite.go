package server

import (
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"runtime"
	"syscall"
)

type DevServer struct {
	Port       int
	Host       string
	PackageMgr string
}

func NewDevServer(port int, host string) *DevServer {
	return &DevServer{
		Port:       port,
		Host:       host,
		PackageMgr: "npm",
	}
}

func (s *DevServer) Start() error {
	fmt.Printf("⚡ Starting nowTS Dev Server...\n")

	cmd := exec.Command("npx", "vite")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start Vite server: %w", err)
	}

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sigChan
		fmt.Println("\n🛑 Shutting down nowTS Dev Server...")
		if runtime.GOOS == "windows" {
			_ = cmd.Process.Kill()
		} else {
			_ = cmd.Process.Signal(syscall.SIGTERM)
		}
		os.Exit(0)
	}()

	return cmd.Wait()
}
