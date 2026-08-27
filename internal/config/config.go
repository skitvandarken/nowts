package config

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

// TargetConfig holds sensitive credentials before encryption
type TargetConfig struct {
	User    string `json:"user"`
	KeyPath string `json:"key_path"`
	WebRoot string `json:"web_root"`
}

// EncryptedDeployConfig is the payload written to .now/deploy.json
type EncryptedDeployConfig struct {
	Target    string `json:"target"`
	Host      string `json:"host"`
	Salt      string `json:"salt"`
	Payload   string `json:"payload"` // AES-GCM encrypted payload (hex)
	CreatedAt string `json:"created_at"`
}

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

// SaveEncryptedConfig serializes and encrypts deployment settings using machine-bound AES-256-GCM
func SaveEncryptedConfig(target string, host string, data TargetConfig) error {
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

// LoadEncryptedConfig reads and decrypts .now/deploy.json
func LoadEncryptedConfig() (*EncryptedDeployConfig, *TargetConfig, error) {
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
