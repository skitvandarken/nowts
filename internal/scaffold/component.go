package scaffold

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

type ComponentGenerator struct {
	Name    string
	Dir     string
	RawName string
}

func NewComponentGenerator(name, targetDir string) *ComponentGenerator {
	cleanName := strings.ToLower(strings.TrimSuffix(name, ".component"))
	return &ComponentGenerator{
		Name:    cleanName,
		Dir:     targetDir,
		RawName: name,
	}
}

// Generate creates the NowTS ergonomic triad: .now.ts, .now.html, .now.css
func (g *ComponentGenerator) Generate() error {
	basePath := filepath.Join(g.Dir, g.Name)
	if err := os.MkdirAll(basePath, 0755); err != nil {
		return fmt.Errorf("failed to create component directory: %w", err)
	}

	tsPath := filepath.Join(basePath, g.Name+".now.ts")
	htmlPath := filepath.Join(basePath, g.Name+".now.html")
	cssPath := filepath.Join(basePath, g.Name+".now.css")

	// 1. Create Component HTML (.now.html)
	htmlContent := fmt.Sprintf(`<!-- %s.now.html -->
<div class="%s-container">
  <h2>%s Component Works!</h2>
</div>
`, g.Name, g.Name, capitalize(g.Name))

	if err := os.WriteFile(htmlPath, []byte(htmlContent), 0644); err != nil {
		return fmt.Errorf("failed to write template file: %w", err)
	}

	// 2. Create Component CSS (.now.css)
	cssContent := fmt.Sprintf(`/* %s.now.css */
.%s-container {
  display: block;
  padding: 1rem;
  box-sizing: border-box;
}
`, g.Name, g.Name)

	if err := os.WriteFile(cssPath, []byte(cssContent), 0644); err != nil {
		return fmt.Errorf("failed to write style file: %w", err)
	}

	// 3. Create Component TypeScript (.now.ts)
	tsTemplate := `// {{.Name}}.now.ts
import { Component } from '@nowts/core';

@Component({
  selector: 'app-{{.Name}}',
  templateUrl: './{{.Name}}.now.html',
  styleUrl: './{{.Name}}.now.css'
})
export class {{.ClassName}}Component {
  title = '{{.Name}} component';

  constructor() {}
}
`
	tmpl, err := template.New("component").Parse(tsTemplate)
	if err != nil {
		return err
	}

	f, err := os.Create(tsPath)
	if err != nil {
		return err
	}
	defer f.Close()

	data := struct {
		Name      string
		ClassName string
	}{
		Name:      g.Name,
		ClassName: capitalize(g.Name),
	}

	if err := tmpl.Execute(f, data); err != nil {
		return err
	}

	fmt.Printf("✨ Scaffolding generated under %s/\n", basePath)
	fmt.Printf("  ├── %s.now.ts\n", g.Name)
	fmt.Printf("  ├── %s.now.html\n", g.Name)
	fmt.Printf("  └── %s.now.css\n", g.Name)

	return nil
}

func capitalize(s string) string {
	if len(s) == 0 {
		return ""
	}
	return strings.ToUpper(s[:1]) + s[1:]
}
