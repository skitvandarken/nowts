package remote

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

// IsValidIPv4 validates IPv4 input and prevents command injection attempts
func IsValidIPv4(ipStr string) bool {
	if strings.ContainsAny(ipStr, " ;&|`$<>(){}[]/\\") {
		return false
	}
	parsedIP := net.ParseIP(ipStr)
	return parsedIP != nil && parsedIP.To4() != nil
}

// DiscoverSSHKey looks for standard SSH keys or prompts the user
func DiscoverSSHKey(reader *bufio.Reader) string {
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

// SecureKeyPermissions ensures the private key has 0600 file mode
func SecureKeyPermissions(keyPath string) error {
	info, err := os.Stat(keyPath)
	if err != nil {
		return err
	}

	if info.Mode().Perm() != 0600 {
		fmt.Printf("🔒 Fixing open key permissions for %s (chmod 600)...\n", keyPath)
		return os.Chmod(keyPath, 0600)
	}
	return nil
}

// TestSSHHandshake executes system SSH to verify server access
func TestSSHHandshake(ip string, user string, keyPath string) error {
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
