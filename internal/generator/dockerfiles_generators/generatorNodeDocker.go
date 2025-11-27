package dockerfiles_generators

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"text/template"

	"github.com/Dancoi/gogen-self-deploy/internal/analyzer"
	"github.com/Dancoi/gogen-self-deploy/internal/generator/util"
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
		util.PrintDockerfile(outPath, b)
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

	// 3) Данные анализа
	rawNodeVersion := "20"
	appPort := ""
	buildScript := "build"
	startCmd := "node dist/index.js"
	useDistRuntime := true
	if analysis != nil && len(analysis.Modules) > 0 {
		for _, m := range analysis.Modules {
			if m.Language == analyzer.LanguageJavaScript || m.Language == analyzer.LanguageTypeScript {
				if v := strings.TrimSpace(m.LanguageVersion); v != "" {
					rawNodeVersion = v
				}
				if p := strings.TrimSpace(m.AppPort); p != "" {
					appPort = p
				}
				break
			}
		}
	}
	nodeVersion := normalizeNodeVersion(rawNodeVersion)
	// Если скрипт build не найден, подменим на пустое выполнение
	if !hasBuildScript(repoRoot) {
		buildScript = ""
	}
	if !hasDistDir(repoRoot) {
		startCmd = "node index.js"
		useDistRuntime = false
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
	util.PrintDockerfile(outPath, buf.Bytes())
	return outPath, nil
}

// normalizeNodeVersion приводит сложные выражения ("20.x 22.x 24.x", ">=18 <21", "^18.17.0") к мажорной версии
func normalizeNodeVersion(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "20"
	}
	re := regexp.MustCompile(`\d{1,2}`)
	nums := re.FindAllString(raw, -1)
	if len(nums) == 0 {
		return "20"
	}
	// Выбираем минимальную подходящую мажорную > 10, но не экзотическую.
	var majors []int
	for _, n := range nums {
		majors = append(majors, atoiSafe(n))
	}
	sort.Ints(majors)
	for _, m := range majors {
		if m >= 14 {
			return fmt.Sprintf("%d", m)
		}
	}
	return nums[0]
}

func atoiSafe(s string) int {
	v := 0
	for _, r := range s {
		if r < '0' || r > '9' {
			return v
		}
		v = v*10 + int(r-'0')
	}
	return v
}

func hasBuildScript(root string) bool {
	p := filepath.Join(root, "package.json")
	b, err := os.ReadFile(p)
	if err != nil {
		return false
	}
	return strings.Contains(string(b), "\"build\"")
}

func hasDistDir(root string) bool {
	info, err := os.Stat(filepath.Join(root, "dist"))
	return err == nil && info.IsDir()
}
