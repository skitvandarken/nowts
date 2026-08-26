package main

import (
	"bufio"
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

// TargetConfig armazena os dados sensíveis antes de encriptar
type TargetConfig struct {
	User    string `json:"user"`
	KeyPath string `json:"key_path"`
	WebRoot string `json:"web_root"`
}

// EncryptedDeployConfig é o payload salvo no arquivo .now/deploy.json
type EncryptedDeployConfig struct {
	Target    string `json:"target"`
	Host      string `json:"host"`
	Salt      string `json:"salt"`
	Payload   string `json:"payload"` // Dados sensíveis encriptados em AES-GCM (hex)
	CreatedAt string `json:"created_at"`
}

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

		if isValidIPv4(input) {
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
	selectedKey := discoverSSHKey(reader)
	if selectedKey == "" {
		fmt.Println("❌ No SSH key specified or found.")
		os.Exit(1)
	}

	// 5. Ensure Key File Permissions (chmod 600)
	if err := secureKeyPermissions(selectedKey); err != nil {
		fmt.Printf("⚠️ Warning: Could not auto-fix key permissions: %v\n", err)
	}

	// 6. Test Handshake via System SSH
	fmt.Printf("\n⚡ Testing SSH handshake with %s@%s:22...\n", sshUser, ip)
	err := testSSHHandshakeStdLib(ip, sshUser, selectedKey)
	if err != nil {
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

	// 8. Encriptar e Salvar Configuração
	targetData := TargetConfig{
		User:    sshUser,
		KeyPath: selectedKey,
		WebRoot: webRootInput,
	}

	if err := saveEncryptedConfig(target, ip, targetData); err != nil {
		fmt.Printf("❌ Failed to save encrypted config: %v\n", err)
		os.Exit(1)
	}

	ensureGitignore()

	fmt.Println("🔒 Configuration encrypted and saved to .now/deploy.json (chmod 600)")
	fmt.Printf("\n🎉 Project successfully initialized! Run 'now run' to deploy.\n")
}

func runDeploy() {
	fmt.Println("📦 Reading deployment manifest...")
	cfg, targetData, err := loadEncryptedConfig()
	if err != nil {
		fmt.Printf("❌ Deployment failed: %v\n", err)
		fmt.Println("💡 Try running 'now init' to re-initialize credentials.")
		os.Exit(1)
	}

	fmt.Printf("🚀 Deploying to %s (%s@%s:%s)...\n", cfg.Target, targetData.User, cfg.Host, targetData.WebRoot)
	// O pipeline de build + sync do 'now run' entra aqui
	fmt.Println("✅ Environment ready for build transfer!")
}

// -----------------------------------------------------------------------------
// ENCRIPTAÇÃO & SEGURANÇA (AES-256-GCM Nativo)
// -----------------------------------------------------------------------------

func getMachineID() string {
	if data, err := os.ReadFile("/etc/machine-id"); err == nil {
		return strings.TrimSpace(string(data))
	}
	hostname, _ := os.Hostname()
	return fmt.Sprintf("%s-%s-%s", hostname, runtime.GOOS, runtime.GOARCH)
}

func deriveKey(salt []byte) []byte {
	machineID := getMachineID()
	hash := sha256.New()
	hash.Write([]byte(machineID))
	hash.Write(salt)
	return hash.Sum(nil)
}

func saveEncryptedConfig(target string, host string, data TargetConfig) error {
	rawBytes, err := json.Marshal(data)
	if err != nil {
		return err
	}

	salt := make([]byte, 16)
	if _, err := io.ReadFull(rand.Reader, salt); err != nil {
		return err
	}

	key := deriveKey(salt)

	block, err := aes.NewCipher(key)
	if err != nil {
		return err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return err
	}

	ciphertext := gcm.Seal(nonce, nonce, rawBytes, nil)

	cfg := EncryptedDeployConfig{
		Target:    target,
		Host:      host,
		Salt:      hex.EncodeToString(salt),
		Payload:   hex.EncodeToString(ciphertext),
		CreatedAt: time.Now().Format(time.RFC3339),
	}

	if err := os.MkdirAll(".now", 0700); err != nil {
		return err
	}

	fileBytes, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}

	filePath := filepath.Join(".now", "deploy.json")
	if err := os.WriteFile(filePath, fileBytes, 0600); err != nil {
		return err
	}

	return os.Chmod(filePath, 0600)
}

func loadEncryptedConfig() (*EncryptedDeployConfig, *TargetConfig, error) {
	filePath := filepath.Join(".now", "deploy.json")
	fileBytes, err := os.ReadFile(filePath)
	if err != nil {
		return nil, nil, fmt.Errorf("could not read .now/deploy.json: %w", err)
	}

	var cfg EncryptedDeployConfig
	if err := json.Unmarshal(fileBytes, &cfg); err != nil {
		return nil, nil, fmt.Errorf("invalid deploy.json format: %w", err)
	}

	salt, err := hex.DecodeString(cfg.Salt)
	if err != nil {
		return nil, nil, fmt.Errorf("invalid salt format")
	}

	ciphertext, err := hex.DecodeString(cfg.Payload)
	if err != nil {
		return nil, nil, fmt.Errorf("invalid payload format")
	}

	key := deriveKey(salt)
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, nil, err
	}

	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return nil, nil, fmt.Errorf("malformed encrypted payload")
	}

	nonce, actualCiphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, actualCiphertext, nil)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to decrypt payload (key mismatch or unauthorized machine)")
	}

	var data TargetConfig
	if err := json.Unmarshal(plaintext, &data); err != nil {
		return nil, nil, fmt.Errorf("corrupted decrypted data: %w", err)
	}

	return &cfg, &data, nil
}

// -----------------------------------------------------------------------------
// UTILITÁRIOS REDE E SISTEMA
// -----------------------------------------------------------------------------

func isValidIPv4(ipStr string) bool {
	if strings.ContainsAny(ipStr, " ;&|`$<>(){}[]/\\") {
		return false
	}
	parsedIP := net.ParseIP(ipStr)
	return parsedIP != nil && parsedIP.To4() != nil
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

func secureKeyPermissions(keyPath string) error {
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

func ensureGitignore() {
	giPath := ".gitignore"
	entry := "\n# nowTS deployment config\n.now/\n"

	content, err := os.ReadFile(giPath)
	if err == nil && strings.Contains(string(content), ".now/") {
		return
	}

	f, err := os.OpenFile(giPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return
	}
	defer f.Close()

	f.WriteString(entry)
}
