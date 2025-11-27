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

// GenerateGoDockerfile рендерит мультистейдж Dockerfile из шаблона
// templates/dockerfiles/go/alpine/Dockerfile_go_multistage.tmpl
// и сохраняет его в gentmp/Dockerfile. Если в репозитории уже есть Dockerfile,
// копирует его вместо рендера. Выводит содержимое в консоль.
func GenerateGoDockerfile(repoRoot string, analysis *analyzer.ProjectAnalysisResult) (string, error) {
	// 1) Целевая папка
	tmpDir := "gentmp"
	if err := os.MkdirAll(tmpDir, 0o755); err != nil {
		return "", fmt.Errorf("mkdir gentmp: %w", err)
	}
	outPath := filepath.Join(tmpDir, "Dockerfile")

	// 2) Если Dockerfile уже есть в репо — копируем
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

	// 3) Иначе рендерим из шаблона мультистейдж
	tplPath := filepath.Join("templates", "dockerfiles", "go", "alpine", "Dockerfile_go_multistage.tmpl")
	raw, err := os.ReadFile(tplPath)
	if err != nil {
		return "", fmt.Errorf("read go dockerfile template: %w", err)
	}
	funcs := template.FuncMap{
		"default": func(def string, val string) string {
			if strings.TrimSpace(val) == "" {
				return def
			}
			return val
		},
		"printf": fmt.Sprintf,
	}
	tpl, err := template.New("go-dockerfile").Funcs(funcs).Option("missingkey=zero").Parse(string(raw))
	if err != nil {
		return "", fmt.Errorf("parse go dockerfile template: %w", err)
	}

	// 4) Данные
	goVersion := "1.22"
	binaryName := "app"
	appPort := ""
	if analysis != nil && len(analysis.Modules) > 0 {
		m := analysis.Modules[0]
		if v := strings.TrimSpace(m.LanguageVersion); v != "" {
			goVersion = v
		}
		rawName := strings.TrimSpace(m.Name)
		if rawName != "" {
			parts := strings.Split(rawName, "/")
			binaryName = sanitizeBinaryName(parts[len(parts)-1])
		}
		appPort = strings.TrimSpace(m.AppPort)
	}
	data := map[string]any{
		"GoVersion":        goVersion,
		"BaseImageBuilder": fmt.Sprintf("golang:%s-alpine", goVersion),
		"BaseImageRuntime": "alpine:3.20",
		"AppWorkdir":       "/app",
		"BinaryName":       binaryName,
		"CGOEnabled":       "0",
		"ExposePort":       appPort,
		"RunTests":         false,
		"LdFlags":          "",
		"Env":              map[string]string{},
		"BuildArgs":        map[string]string{},
	}

	var buf bytes.Buffer
	if err := tpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("render go dockerfile: %w", err)
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

func sanitizeBinaryName(name string) string {
	s := strings.ToLower(name)
	var b strings.Builder
	for _, r := range s {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' || r == '_' {
			b.WriteRune(r)
		} else {
			b.WriteRune('-')
		}
	}
	return strings.Trim(b.String(), "-")
}
