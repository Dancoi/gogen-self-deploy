package dockerfiles_generators

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"text/template"

	"github.com/Dancoi/gogen-self-deploy/internal/analyzer"
)

func GenerateNodeDockerfile(repoRoot string, analysis *analyzer.ProjectAnalysisResult) (string, error) {
	tmpDir := "gentmp"
	if err := os.MkdirAll(tmpDir, 0o755); err != nil {
		return "", fmt.Errorf("mkdir gentmp: %w", err)
	}
	outPath := filepath.Join(tmpDir, "Dockerfile")

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

	nodeVersion := "20"
	appPort := ""
	buildScript := "build"
	startCmd := "node dist/index.js"
	useDistRuntime := true
	if analysis != nil && len(analysis.Modules) > 0 {
		for _, m := range analysis.Modules {
			if m.Language == analyzer.LanguageJavaScript || m.Language == analyzer.LanguageTypeScript {
				if v := strings.TrimSpace(m.LanguageVersion); v != "" {
					// parse version string and extract the highest major version number
					if major := extractMajorVersion(v); major != "" {
						nodeVersion = major
					} else {
						nodeVersion = v
					}
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

// extractMajorVersion finds integers in a version expression and returns the maximum (major) version as string.
// Examples:
//
//	">=20.0.0 <=24.x.x" -> "24"
//	"20.x || 22.x || 24.x" -> "24"
//	"^16.3.0" -> "16"
func extractMajorVersion(s string) string {
	// find all number sequences
	re := regexp.MustCompile(`\d+`)
	matches := re.FindAllString(s, -1)
	if len(matches) == 0 {
		return ""
	}
	max := -1
	for _, m := range matches {
		n, err := strconv.Atoi(m)
		if err != nil {
			continue
		}
		if n > max {
			max = n
		}
	}
	if max < 0 {
		return ""
	}
	return strconv.Itoa(max)
}

// ResolveNodeMajor is an exported helper to resolve major Node version from a version expression.
// It wraps the internal parser and is intended for quick tests and external callers.
func ResolveNodeMajor(s string) string {
	return extractMajorVersion(s)
}

func printDockerfile(path string, content []byte) {
	fmt.Println("Saved Dockerfile to:", path)
	fmt.Println("----- Dockerfile -----")
	fmt.Println(string(content))
	fmt.Println("----- end -----")
}
