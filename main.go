package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"nowts/internal/config"
	"nowts/internal/dev"
	"nowts/internal/remote"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: now <command> [options]")
		fmt.Println("Available commands: init, serve, run, dev, build")
		os.Exit(1)
	}

	command := os.Args[1]

	switch command {
	case "init":
		runInit()
	case "serve", "dev":
		fmt.Println("🚀 Starting development server...")
	case "run", "build":
		runDeploy()
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

	// 2. IPv4 Input
	var ip string
	for {
		fmt.Print("Server IPv4 Address (e.g., 192.168.1.100): ")
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)

		if remote.IsValidIPv4(input) {
			ip = input
			break
		}
		fmt.Println("❌ Invalid IPv4 address. Please try again.")
	}

	// 3. Username Input
	fmt.Print("SSH Username (default: ubuntu): ")
	sshUser, _ := reader.ReadString('\n')
	sshUser = strings.TrimSpace(sshUser)
	if sshUser == "" {
		sshUser = "ubuntu"
	}

	// 4. Auto-Discover or Prompt Key
	selectedKey := remote.DiscoverSSHKey(reader)
	if selectedKey == "" {
		fmt.Println("❌ No SSH key specified or found.")
		os.Exit(1)
	}

	// 5. Ensure Key File Permissions (chmod 600)
	if err := remote.SecureKeyPermissions(selectedKey); err != nil {
		fmt.Printf("⚠️ Warning: Could not auto-fix key permissions: %v\n", err)
	}

	// 6. Test Handshake via System SSH
	fmt.Printf("\n⚡ Testing SSH handshake with %s@%s:22...\n", sshUser, ip)
	if err := remote.TestSSHHandshake(ip, sshUser, selectedKey); err != nil {
		fmt.Printf("❌ Handshake failed: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("✅ SSH Handshake successful!")

	// 7. Define Remote Web Root
	defaultWebRoot := "/var/www/html"
	if target == "cpanel" {
		defaultWebRoot = "/public_html"
	}
	fmt.Printf("Remote Web Root directory (default: %s): ", defaultWebRoot)
	webRootInput, _ := reader.ReadString('\n')
	webRootInput = strings.TrimSpace(webRootInput)
	if webRootInput == "" {
		webRootInput = defaultWebRoot
	}

	// 8. Encrypt and Save Config
	targetData := config.TargetConfig{
		User:    sshUser,
		KeyPath: selectedKey,
		WebRoot: webRootInput,
	}

	if err := config.SaveEncryptedConfig(target, ip, targetData); err != nil {
		fmt.Printf("❌ Failed to save encrypted config: %v\n", err)
		os.Exit(1)
	}

	dev.EnsureGitignore()

	fmt.Println("🔒 Configuration encrypted and saved to .now/deploy.json (chmod 600)")
	fmt.Printf("\n🎉 Project successfully initialized! Run 'now run' to deploy.\n")
}

func runDeploy() {
	fmt.Println("📦 Reading deployment manifest...")
	cfg, targetData, err := config.LoadEncryptedConfig()
	if err != nil {
		fmt.Printf("❌ Deployment failed: %v\n", err)
		fmt.Println("💡 Try running 'now init' to re-initialize credentials.")
		os.Exit(1)
	}

	fmt.Printf("🚀 Deploying to %s (%s@%s:%s)...\n", cfg.Target, targetData.User, cfg.Host, targetData.WebRoot)
	// Build + sync pipeline entrypoint
	fmt.Println("✅ Environment ready for build transfer!")
}
