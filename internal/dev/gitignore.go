package dev

import (
	"os"
	"strings"
)

// EnsureGitignore ensures the binary and generated source code are excluded from git
func EnsureGitignore() {
	gitignorePath := ".gitignore"
	entries := []string{
		"now",
		"now.exe",
		"src/",
		".now/",
	}

	content, err := os.ReadFile(gitignorePath)
	if os.IsNotExist(err) {
		_ = os.WriteFile(gitignorePath, []byte(strings.Join(entries, "\n")+"\n"), 0644)
		return
	}

	existing := string(content)
	var toAppend []string

	for _, entry := range entries {
		if !strings.Contains(existing, entry) {
			toAppend = append(toAppend, entry)
		}
	}

	if len(toAppend) > 0 {
		f, err := os.OpenFile(gitignorePath, os.O_APPEND|os.O_WRONLY, 0644)
		if err == nil {
			defer f.Close()
			for _, entry := range toAppend {
				_, _ = f.WriteString("\n" + entry)
			}
		}
	}
}
