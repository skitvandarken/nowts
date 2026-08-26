package main

import (
	"bufio"
	"context"
	"fmt"
	"net"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"strings"
	"time"
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

	// 2. IPv4 Input with Sanitization
	var ip string
	for {
		fmt.Print("Server IPv4 Address (e.g., 192.168.1.100): ")
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)

		if isValidIPv4(input) {
			ip = input
			break
		}
		fmt.Println("❌ Invalid IPv4 address. Please try again.")
	}

	// 3. Username Input
	fmt.Print("SSH Username (default: root): ")
	user, _ := reader.ReadString('\n')
	user = strings.TrimSpace(user)
	if user == "" {
		user = "root"
	}

	// 4. Auto-Discover Key via Standard Library
	selectedKey := discoverSSHKey(reader)
	if selectedKey == "" {
		fmt.Println("❌ No SSH key specified or found.")
		os.Exit(1)
	}

	// Auto-correct SSH key permissions (0600 required by SSH)
	if err := enforceKeyPermissions(selectedKey); err != nil {
		fmt.Printf("⚠️ Warning: Could not set key permissions: %v\n", err)
	}

	// 5. Test Handshake via System SSH
	fmt.Printf("\n⚡ Testing SSH handshake with %s@%s:22...\n", user, ip)
	err := testSSHHandshakeStdLib(ip, user, selectedKey)
	if err != nil {
		fmt.Printf("❌ Handshake failed: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("✅ SSH Handshake successful!")
	fmt.Printf("\n🎉 Project successfully initialized for %s (%s@%s) using key: %s\n", target, user, ip, selectedKey)
}

func isValidIPv4(ipStr string) bool {
	if strings.ContainsAny(ipStr, " ;&|`$<>(){}[]/\\") {
		return false
	}
	parsedIP := net.ParseIP(ipStr)
	return parsedIP != nil && parsedIP.To4() != nil
}

func enforceKeyPermissions(keyPath string) error {
	info, err := os.Stat(keyPath)
	if err != nil {
		return err
	}

	// If permissions are broader than 0600 (owner read/write only)
	if info.Mode().Perm() != 0600 {
		fmt.Printf("🔒 Adjusting permissions for %s to 0600 (required by SSH)...\n", keyPath)
		return os.Chmod(keyPath, 0600)
	}
	return nil
}

func discoverSSHKey(reader *bufio.Reader) string {
	usr, err := user.Current()
	if err != nil {
		home := os.Getenv("HOME")
		if home == "" {
			return promptCustomKey(reader)
		}
		usr = &user.User{HomeDir: home}
	}

	candidates := []string{
		filepath.Join(usr.HomeDir, ".ssh", "id_ed25519"),
		filepath.Join(usr.HomeDir, ".ssh", "id_rsa"),
	}

	for _, path := range candidates {
		if _, err := os.Stat(path); err == nil {
			fmt.Printf("🔑 Discovered default SSH key: %s\n", path)
			return path
		}
	}

	return promptCustomKey(reader)
}

func promptCustomKey(reader *bufio.Reader) string {
	fmt.Print("Enter path to your SSH private key: ")
	input, _ := reader.ReadString('\n')
	return strings.TrimSpace(input)
}

func testSSHHandshakeStdLib(ip string, user string, keyPath string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	args := []string{
		"-i", keyPath,
		"-o", "BatchMode=yes",
		"-o", "ConnectTimeout=5",
		"-o", "StrictHostKeyChecking=accept-new",
		fmt.Sprintf("%s@%s", user, ip),
		"exit 0",
	}

	cmd := exec.CommandContext(ctx, "ssh", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return fmt.Errorf("connection timed out after 5 seconds")
		}
		cleanErr := strings.TrimSpace(string(output))
		if cleanErr != "" {
			return fmt.Errorf("%s", cleanErr)
		}
		return err
	}

	return nil
}
