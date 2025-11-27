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

// GeneratePythonDockerfile генерирует (или копирует существующий) мультистейдж Dockerfile для Python
// Использует шаблон templates/dockerfiles/python/slim/Dockerfile_python_multistage.tmpl
// Сохраняет в gentmp/Dockerfile и печатает содержимое.
func GeneratePythonDockerfile(repoRoot string, analysis *analyzer.ProjectAnalysisResult) (string, error) {
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
		fmt.Println("Saved Dockerfile to:", outPath)
		fmt.Println("----- Dockerfile -----")
		fmt.Println(string(b))
		fmt.Println("----- end -----")
		return outPath, nil
	}

	// 2) Если анализатор указал путь к Dockerfile
	if analysis != nil {
		for _, m := range analysis.Modules {
			if m.Language == analyzer.LanguagePython && strings.TrimSpace(m.DockerfilePath) != "" {
				abs := m.DockerfilePath
				if !filepath.IsAbs(abs) {
					abs = filepath.Join(repoRoot, abs)
				}
				if fi, err := os.Stat(abs); err == nil && !fi.IsDir() {
					b, err := os.ReadFile(abs)
					if err != nil {
						return "", err
					}
					if err := os.WriteFile(outPath, b, 0o644); err != nil {
						return "", err
					}
					fmt.Println("Saved Dockerfile to:", outPath)
					fmt.Println("----- Dockerfile -----")
					fmt.Println(string(b))
					fmt.Println("----- end -----")
					return outPath, nil
				}
			}
		}
	}

	// 3) Рендер из шаблона (multistage)
	tplPath := filepath.Join("templates", "dockerfiles", "python", "slim", "Dockerfile_python_multistage.tmpl")
	raw, err := os.ReadFile(tplPath)
	if err != nil {
		return "", fmt.Errorf("read python dockerfile template: %w", err)
	}
	funcs := template.FuncMap{
		"default": func(def string, val string) string {
			if strings.TrimSpace(val) == "" {
				return def
			}
			return val
		},
	}
	tpl, err := template.New("python-dockerfile").Funcs(funcs).Option("missingkey=zero").Parse(string(raw))
	if err != nil {
		return "", fmt.Errorf("parse python dockerfile template: %w", err)
	}

	pyVersion := "3.12"
	appPort := ""
	if analysis != nil {
		for _, m := range analysis.Modules {
			if m.Language == analyzer.LanguagePython {
				if v := strings.TrimSpace(m.LanguageVersion); v != "" {
					pyVersion = v
				}
				if p := strings.TrimSpace(m.AppPort); p != "" {
					appPort = p
				}
				break
			}
		}
	}
	data := map[string]any{
		"BaseImageBuilder": fmt.Sprintf("python:%s-slim", pyVersion),
		"BaseImageRuntime": fmt.Sprintf("python:%s-slim", pyVersion),
		"AppWorkdir":       "/app",
		"RequirementsFile": "requirements.txt",
		"UseVenv":          "true",
		"Env":              map[string]string{},
		"BuildArgs":        map[string]string{},
		"RunTests":         false,
		"Entrypoint":       []string{"python", "-m", "app"},
		"ExposePort":       appPort,
	}
	var buf bytes.Buffer
	if err := tpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("render python dockerfile: %w", err)
	}
	if err := os.WriteFile(outPath, buf.Bytes(), 0o644); err != nil {
		return "", err
	}
	fmt.Println("Saved Dockerfile to:", outPath)
	fmt.Println("----- Dockerfile -----")
	fmt.Println(buf.String())
	fmt.Println("----- end -----")
	return outPath, nil
}
