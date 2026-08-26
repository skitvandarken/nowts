package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: now <command> [options]")
		fmt.Println("Available commands: init, serve, run, g")
		os.Exit(1)
	}

	command := os.Args[1]

	switch command {
	case "init":
		runInit()
	case "serve", "dev":
		fmt.Println("🚀 Starting development server...")
	case "run", "build":
		fmt.Println("📦 Compiling and deploying to production...")
	default:
		fmt.Printf("Command '%s' not recognized.\n", command)
	}
}

func runInit() {
	reader := bufio.NewReader(os.Stdin)

	fmt.Println("--- nowTS Configuration ---")

	// 1. Target Selection
	fmt.Println("Select the deployment target:")
	fmt.Println("1) Linux VPS (SSH/Caddy/Nginx)")
	fmt.Println("2) cPanel / Shared Hosting (SFTP)")
	fmt.Print("Option (1 or 2): ")

	targetInput, _ := reader.ReadString('\n')
	targetInput = strings.TrimSpace(targetInput)

	target := "vps"
	if targetInput == "2" {
		target = "cpanel"
	}

	// 2. Host Input
	fmt.Print("Server Host / Domain (e.g., app.mysite.com): ")
	host, _ := reader.ReadString('\n')
	host = strings.TrimSpace(host)

	fmt.Printf("\n✅ Project configured for %s on host %s!\n", target, host)
}
