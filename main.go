package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"os"
	"strings"

	"nowts/internal/compiler"
	"nowts/internal/config"
	"nowts/internal/dev"
	"nowts/internal/remote"
	"nowts/internal/scaffold"
	"nowts/internal/server"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	// Subcommands
	initCmd := flag.NewFlagSet("init", flag.ExitOnError)
	devInit := initCmd.Bool("dev", false, "Quick initialize dev mode environment")

	generateCmd := flag.NewFlagSet("generate", flag.ExitOnError)

	switch os.Args[1] {
	case "init":
		initCmd.Parse(os.Args[2:])
		if *devInit {
			runDevInit()
		} else {
			runInit()
		}

	case "g", "generate":
		generateCmd.Parse(os.Args[2:])
		
		// generateCmd.Arg(0) -> "component" or "c"
		// generateCmd.Arg(1) -> actual name (e.g., "header")
		targetType := generateCmd.Arg(0)
		targetName := generateCmd.Arg(1)

		switch targetType {
		case "component", "c":
			if targetName == "" {
				fmt.Println("❌ Error: Missing component name.")
				fmt.Println("Usage: now generate component <name>")
				os.Exit(1)
			}
			gen := scaffold.NewComponentGenerator(targetName, "src/components")
			if err := gen.Generate(); err != nil {
				fmt.Printf("❌ Scaffolding failed: %v\n", err)
				os.Exit(1)
			}
		default:
			fmt.Printf("Unknown generator target '%s'. Supported: component (c)\n", targetType)
			os.Exit(1)
		}

	case "serve", "dev":
		runServe()

	case "run", "build":
		runDeploy()

	default:
		fmt.Printf("Command '%s' not recognized.\n", os.Args[1])
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("Usage: now <command> [options]")
	fmt.Println("\nCommands:")
	fmt.Println("  init [--dev]              Initialize configuration")
	fmt.Println("  generate component <name> Generate NowTS component triad (.now.ts, .now.html, .now.css)")
	fmt.Println("  serve / dev               Run development server with tsgo type check")
	fmt.Println("  run / build               Type check and deploy build")
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
