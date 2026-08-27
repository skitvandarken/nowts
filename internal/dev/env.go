package dev

import (
	"os"
	"strings"
)

func EnsureGitignore() {
	giPath := ".gitignore"
	entry := "\n# nowTS local configs & secrets\n.now/\n*.key\n"

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
