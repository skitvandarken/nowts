package compiler

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"time"
)

type Compiler struct {
	BinaryName string
}

func NewCompiler() *Compiler {
	return &Compiler{
		BinaryName: "tsgo",
	}
}

func (c *Compiler) EnsureInstalled() error {
	_, err := exec.LookPath(c.BinaryName)
	if err == nil {
		return nil
	}

	fmt.Printf("⚠️ %s not found in PATH. Installing via go install...\n", c.BinaryName)
	cmd := exec.Command("go", "install", "github.com/microsoft/tsgo/cmd/tsgo@latest")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func (c *Compiler) CheckTypes(ctx context.Context) error {
	fmt.Println("⚡ Running instant Go-powered TypeScript type-check...")
	start := time.Now()
	
	cmd := exec.CommandContext(ctx, c.BinaryName, "--noEmit")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("type check failed")
	}

	fmt.Printf("✨ Type check passed in %v!\n", time.Since(start))
	return nil
}
