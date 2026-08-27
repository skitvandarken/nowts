package compiler

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"time"
)

// TSConfig represents the standard tsconfig.json schema
type TSConfig struct {
	CompilerOptions CompilerOptions `json:"compilerOptions"`
	Include         []string        `json:"include"`
	Exclude         []string        `json:"exclude,omitempty"`
}

type CompilerOptions struct {
	Target           string `json:"target"`
	Module           string `json:"module"`
	ModuleResolution string `json:"moduleResolution"`
	Strict           bool   `json:"strict"`
	JSX              string `json:"jsx"`
	SkipLibCheck     bool   `json:"skipLibCheck"`
	IsolatedModules  bool   `json:"isolatedModules"`
	NoEmit           bool   `json:"noEmit"`
}

type Compiler struct {
	BinaryPath string
	OS         string
}

func NewCompiler() *Compiler {
	return &Compiler{
		OS: runtime.GOOS,
	}
}

// EnsureInstalled guarantees tsgo exists in PATH or installs it directly via Go
func (c *Compiler) EnsureInstalled() error {
	// Look for native tsgo binary
	path, err := exec.LookPath("tsgo")
	if err == nil && path != "" {
		c.BinaryPath = path
		return nil
	}

	// Look for typescript-go binary alternate name
	path, err = exec.LookPath("typescript-go")
	if err == nil && path != "" {
		c.BinaryPath = path
		return nil
	}

	// Attempt native Go installation if missing
	fmt.Println("⚠️ tsgo not found in PATH. Installing Go-native TypeScript compiler...")
	cmd := exec.Command("go", "install", "github.com/microsoft/tsgo/cmd/tsgo@latest")
	if err := cmd.Run(); err != nil {
		// Fallback check in $GOPATH/bin
		gopath := os.Getenv("GOPATH")
		if gopath == "" {
			gopath = filepath.Join(os.Getenv("HOME"), "go")
		}
		candidate := filepath.Join(gopath, "bin", "tsgo")
		if _, err := os.Stat(candidate); err == nil {
			c.BinaryPath = candidate
			return nil
		}
		return fmt.Errorf("tsgo installation failed: %w (ensure $GOPATH/bin is in your PATH)", err)
	}

	c.BinaryPath = "tsgo"
	return nil
}

// CheckTypes parses/ensures tsconfig.json using json package and runs tsgo
func (c *Compiler) CheckTypes(ctx context.Context) error {
	if err := c.ensureTSConfig(); err != nil {
		return fmt.Errorf("failed to process tsconfig.json: %w", err)
	}

	stopSpinner := startSpinner("Running tsgo type-check...")
	start := time.Now()

	cmd := exec.CommandContext(ctx, c.BinaryPath, "--noEmit")
	output, err := cmd.CombinedOutput()
	stopSpinner()

	if err != nil {
		fmt.Printf("\r❌ TypeScript type-check failed (%v):\n", time.Since(start).Round(time.Millisecond))
		fmt.Println(string(output))
		return fmt.Errorf("type check failed")
	}

	fmt.Printf("\r✨ Type-check passed via tsgo (%v)\n", time.Since(start).Round(time.Millisecond))
	return nil
}

// Ensure tsconfig.json exists or parse/merge default values using encoding/json
func (c *Compiler) ensureTSConfig() error {
	configPath := "tsconfig.json"

	defaultConfig := TSConfig{
		CompilerOptions: CompilerOptions{
			Target:           "ESNext",
			Module:           "ESNext",
			ModuleResolution: "bundler",
			Strict:           true,
			JSX:              "preserve",
			SkipLibCheck:     true,
			IsolatedModules:  true,
			NoEmit:           true,
		},
		Include: []string{"src/**/*"},
	}

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		data, err := json.MarshalIndent(defaultConfig, "", "  ")
		if err != nil {
			return err
		}
		return os.WriteFile(configPath, data, 0644)
	}

	// Parse existing tsconfig.json to validate standard structure
	content, err := os.ReadFile(configPath)
	if err != nil {
		return err
	}

	var parsedConfig TSConfig
	if err := json.Unmarshal(content, &parsedConfig); err != nil {
		// Write out defaults if existing file is invalid JSON
		data, _ := json.MarshalIndent(defaultConfig, "", "  ")
		return os.WriteFile(configPath, data, 0644)
	}

	return nil
}

func startSpinner(message string) func() {
	done := make(chan bool)
	go func() {
		frames := []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
		i := 0
		for {
			select {
			case <-done:
				return
			default:
				fmt.Printf("\r\033[K%s %s", frames[i%len(frames)], message)
				time.Sleep(80 * time.Millisecond)
				i++
			}
		}
	}()

	return func() {
		done <- true
		time.Sleep(10 * time.Millisecond)
		fmt.Print("\r\033[K")
	}
}
