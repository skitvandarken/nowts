package dev

import (
	"os"
	"strings"
)

// EnsureGitignore ensures that the .now/ directory is excluded from version control
func EnsureGitignore() {
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
