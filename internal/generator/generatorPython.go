package generator

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"
	"time"

	"github.com/Dancoi/gogen-self-deploy/internal/analyzer"
)

type pythonReport struct {
	LanguageVersion string
	AppPort         string
}

type pythonOpt struct {
	SonarHost       string
	RegistryProject string
}

type pythonTplData struct {
	Report             pythonReport
	Opt                pythonOpt
	Now                string
	CI_COMMIT_REF_SLUG string
}

// GeneratePythonPipeline рендерит GitLab CI из python-темплейта, создаёт 'gentmp',
// сохраняет 'gentmp/.gitlab-ci.yml' и печатает его.
func GeneratePythonPipeline(repoName string, analysis *analyzer.ProjectAnalysisResult) error {
	// 1) Папка для вывода
	tmpDir := "gentmp"
	if err := os.MkdirAll(tmpDir, 0o755); err != nil {
		return fmt.Errorf("mkdir gentmp: %w", err)
	}

	// 2) Чтение шаблона
	tplPath := filepath.Join("templates", "gitlab", "pipelines", "python.gitlab-ci.yml.tmpl")
	raw, err := os.ReadFile(tplPath)
	if err != nil {
		return fmt.Errorf("read template: %w", err)
	}

	// 3) Данные из анализа
	report := pythonReport{
		LanguageVersion: "",
		AppPort:         "8000",
	}
	if m := firstPythonModule(analysis); m != nil {
		if v := strings.TrimSpace(m.LanguageVersion); v != "" {
			report.LanguageVersion = v
		} else if v := strings.TrimSpace(m.FrameworkVersion); v != "" {
			// запасной вариант, если версия ошибочно попала во фреймворк
			report.LanguageVersion = v
		}
		if p := strings.TrimSpace(m.AppPort); p != "" {
			report.AppPort = p
		}
	}

	// 4) Опции — можно пробрасывать через ENV
	opts := pythonOpt{
		SonarHost:       getenvDefault("SONAR_HOST_URL", ""),
		RegistryProject: getenvDefault("REGISTRY_PROJECT", ""),
	}

	data := pythonTplData{
		Report:             report,
		Opt:                opts,
		Now:                time.Now().Format(time.RFC3339),
		CI_COMMIT_REF_SLUG: "$CI_COMMIT_REF_SLUG",
	}

	// 5) Рендер
	tpl, err := template.New("python-ci").Option("missingkey=zero").Parse(string(raw))
	if err != nil {
		return fmt.Errorf("parse template: %w", err)
	}
	var buf bytes.Buffer
	if err := tpl.Execute(&buf, data); err != nil {
		return fmt.Errorf("execute template: %w", err)
	}

	// 6) Сохранение и вывод
	outPath := filepath.Join(tmpDir, ".gitlab-ci.yml")
	if err := os.WriteFile(outPath, buf.Bytes(), 0o644); err != nil {
		return fmt.Errorf("write .gitlab-ci.yml: %w", err)
	}

	fmt.Println("----- .gitlab-ci.yml -----")
	fmt.Print(buf.String())
	fmt.Println("----- end -----")
	fmt.Println("Saved to:", outPath)
	return nil
}

func firstPythonModule(analysis *analyzer.ProjectAnalysisResult) *analyzer.ProjectModule {
	if analysis == nil {
		return nil
	}
	for _, m := range analysis.Modules {
		if m.Language == analyzer.LanguagePython {
			return m
		}
	}
	return nil
}

func getenvDefault(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
