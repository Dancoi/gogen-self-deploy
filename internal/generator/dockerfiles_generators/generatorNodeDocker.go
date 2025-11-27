package dockerfiles_generators

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/Dancoi/gogen-self-deploy/internal/analyzer"
)

// GenerateNodeDockerfile генерирует (или копирует существующий) мультистейдж Dockerfile для Node/TS
// Использует шаблон templates/dockerfiles/node/alpine/Dockerfile_node_multistage.tmpl
// Сохраняет в gentmp/Dockerfile и печатает содержимое.
func GenerateNodeDockerfile(repoRoot string, analysis *analyzer.ProjectAnalysisResult) (string, error) {
	tmpDir := "gentmp"
	if err := os.MkdirAll(tmpDir, 0o755); err != nil {
		return "", fmt.Errorf("mkdir gentmp: %w", err)
	}
	outPath := filepath.Join(tmpDir, "Dockerfile")

	// 1) Если есть существующий Dockerfile в корне
	existing := filepath.Join(repoRoot, "Dockerfile")
	if fi, err := os.Stat(existing); err == nil && !fi.IsDir() {
		b, err := os.ReadFile(existing)
		if err != nil {
			return "", err
		}
		if err := os.WriteFile(outPath, b, 0o644); err != nil {
			return "", err
		}
		printDockerfile(outPath, b)
		return outPath, nil
	}

	// 2) Шаблон
	tplPath := filepath.Join("templates", "dockerfiles", "node", "alpine", "Dockerfile_node_multistage.tmpl")
	raw, err := os.ReadFile(tplPath)
	if err != nil {
		return "", fmt.Errorf("read node dockerfile template: %w", err)
	}
	funcs := template.FuncMap{
		"default": func(def string, val string) string {
			if strings.TrimSpace(val) == "" {
				return def
			}
			return val
		},
		"upper": strings.ToUpper,
	}
	tpl, err := template.New("node-dockerfile").Funcs(funcs).Option("missingkey=zero").Parse(string(raw))
	if err != nil {
		return "", fmt.Errorf("parse node dockerfile template: %w", err)
	}

	// 3) Данные анализа
	nodeVersion := "20"
	appPort := ""
	buildScript := "build"
	startCmd := "node dist/index.js"
	useDistRuntime := true
	if analysis != nil && len(analysis.Modules) > 0 {
		for _, m := range analysis.Modules {
			if m.Language == analyzer.LanguageJavaScript || m.Language == analyzer.LanguageTypeScript {
				if v := strings.TrimSpace(m.LanguageVersion); v != "" {
					nodeVersion = v
				}
				if p := strings.TrimSpace(m.AppPort); p != "" {
					appPort = p
				}
				break
			}
		}
	}

	data := map[string]any{
		"BaseImageBuilder": fmt.Sprintf("node:%s-alpine", nodeVersion),
		"BaseImageRuntime": fmt.Sprintf("node:%s-alpine", nodeVersion),
		"AppWorkdir":       "/app",
		"UseDistRuntime":   useDistRuntime,
		"BuildArgs":        map[string]string{},
		"Env":              map[string]string{},
		"BuildScript":      buildScript,
		"StartCommand":     startCmd,
		"ExposePort":       appPort,
	}

	var buf bytes.Buffer
	if err := tpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("render node dockerfile: %w", err)
	}
	if err := os.WriteFile(outPath, buf.Bytes(), 0o644); err != nil {
		return "", err
	}
	printDockerfile(outPath, buf.Bytes())
	return outPath, nil
}

func printDockerfile(path string, content []byte) {
	fmt.Println("Saved Dockerfile to:", path)
	fmt.Println("----- Dockerfile -----")
	fmt.Println(string(content))
	fmt.Println("----- end -----")
}
