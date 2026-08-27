package dev

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

// Environment manages compiler requirements and project scaffolding
type Environment struct {
	CompilerName string
	Dirs         []string
}

func NewEnvironment() *Environment {
	return &Environment{
		CompilerName: "esbuild",
		Dirs: []string{
			"src",
			"src/components",
			"public",
			"dist",
		},
	}
}

// EnsureToolchain checks if the Go-based TS compiler exists or installs it
func (e *Environment) EnsureToolchain() error {
	fmt.Printf("🔍 Checking for TypeScript compiler (%s)...\n", e.CompilerName)
	_, err := exec.LookPath(e.CompilerName)
	if err != nil {
		fmt.Printf("⚠️ %s not found. Installing via go install...\n", e.CompilerName)

		cmd := exec.Command("go", "install", "github.com/evanw/esbuild/cmd/esbuild@latest")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to install %s: %w", e.CompilerName, err)
		}
		fmt.Printf("✅ Installed %s!\n", e.CompilerName)
	} else {
		fmt.Printf("✅ Toolchain verified: %s\n", e.CompilerName)
	}
	return nil
}

// ScaffoldDirs creates the development directory structure and entrypoint
func (e *Environment) ScaffoldDirs() error {
	fmt.Println("📁 Scaffolding project structure...")
	for _, dir := range e.Dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed creating directory %s: %w", dir, err)
		}
	}

	entryFile := filepath.Join("src", "index.ts")
	if _, err := os.Stat(entryFile); os.IsNotExist(err) {
		sample := []byte("// nowTS entrypoint\nconsole.log('App ready');\n")
		if err := os.WriteFile(entryFile, sample, 0644); err != nil {
			return fmt.Errorf("failed writing %s: %w", entryFile, err)
		}
		fmt.Printf("📝 Created %s\n", entryFile)
	}
	return nil
}
