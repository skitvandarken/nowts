package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	"nowts/internal/compiler"
	"nowts/internal/config"
	"nowts/internal/dev"
	"nowts/internal/remote"
	"nowts/internal/server"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: now <command> [options]")
		fmt.Println("Commands: init [--dev], serve, run")
		os.Exit(1)
	}

	command := os.Args[1]

	switch command {
	case "init":
		if len(os.Args) > 2 && os.Args[2] == "--dev" {
			runDevInit()
		} else {
			runInit()
		}
	case "serve", "dev":
		runServe()
	case "run", "build":
		runDeploy()
	default:
		fmt.Printf("Command '%s' not recognized.\n", command)
		os.Exit(1)
	}
}

func runDevInit() {
	fmt.Println("🔧 Quick-creating local development environment (.now/dev.json)...")
	dev.EnsureGitignore()
	reader := bufio.NewReader(os.Stdin)

	fmt.Print("Enter Dev IP/Host: ")
	host, _ := reader.ReadString('\n')
	host = strings.TrimSpace(host)

	fmt.Print("Enter SSH User (default: root): ")
	user, _ := reader.ReadString('\n')
	user = strings.TrimSpace(user)
	if user == "" {
		user = "root"
	}

	keyPath := remote.DiscoverSSHKey(reader)
	_ = remote.SecureKeyPermissions(keyPath)

	cfg := config.DevConfig{
		Host:    host,
		User:    user,
		KeyPath: keyPath,
		WebRoot: "/var/www/html",
	}

	if err := config.SaveDevConfig(cfg); err != nil {
		fmt.Printf("❌ Failed to save dev config: %v\n", err)
		return
	}

	fmt.Println("⚡ Saved dev credentials to .now/dev.json!")
}

func runInit() {
	dev.EnsureGitignore()
	reader := bufio.NewReader(os.Stdin)

	fmt.Print("Enter Target IP/Host: ")
	host, _ := reader.ReadString('\n')
	host = strings.TrimSpace(host)

	fmt.Print("Enter SSH User (default: root): ")
	user, _ := reader.ReadString('\n')
	user = strings.TrimSpace(user)
	if user == "" {
		user = "root"
	}

	keyPath := remote.DiscoverSSHKey(reader)
	_ = remote.SecureKeyPermissions(keyPath)

	cfg := config.DevConfig{
		Host:    host,
		User:    user,
		KeyPath: keyPath,
		WebRoot: "/var/www/html",
	}

	if err := config.SaveDevConfig(cfg); err != nil {
		fmt.Printf("❌ Failed: %v\n", err)
		return
	}

	fmt.Println("✅ Saved credentials to .now/dev.json!")
}

func runServe() {
	dev.EnsureGitignore()

	tsCompiler := compiler.NewCompiler()
	if err := tsCompiler.EnsureInstalled(); err == nil {
		ctx := context.Background()
		_ = tsCompiler.CheckTypes(ctx)
	}

	srv := server.NewDevServer(5173, "localhost")
	if err := srv.Start(); err != nil {
		fmt.Printf("❌ Server error: %v\n", err)
		os.Exit(1)
	}
}

func runDeploy() {
	devCfg, err := config.LoadDevConfig()
	if err != nil {
		fmt.Printf("❌ Credentials missing: %v\nRun 'now init --dev' to create .now/dev.json\n", err)
		os.Exit(1)
	}

	fmt.Printf("🚀 Deploying to dev target %s@%s (Key: %s)...\n", devCfg.User, devCfg.Host, devCfg.KeyPath)
}
