package pipelines_generators

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"
	"time"

	"github.com/Dancoi/gogen-self-deploy/internal/analyzer"
	"github.com/Dancoi/gogen-self-deploy/internal/generator/dockerfiles_generators"
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
// генерирует/копирует Dockerfile в gentmp/Dockerfile, сохраняет 'gentmp/.gitlab-ci.yml' и печатает его.
func GeneratePythonPipeline(analysis *analyzer.ProjectAnalysisResult) error {
	// 0) Dockerfile
	if _, err := dockerfiles_generators.GeneratePythonDockerfile(filepath.Join("gentmp", ".."), analysis); err != nil {
		// repoRoot нам не передают; генератор Python Dockerfile сам копирует/рендерит в gentmp/Dockerfile из шаблона
		// Используем шаблонный путь
		_ = err // игнорируем, если шаблон не найден — пайплайн всё равно сгенерируется
	}

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
	report := pythonReport{LanguageVersion: "", AppPort: "8000"}
	if m := firstPythonModule(analysis); m != nil {
		if v := strings.TrimSpace(m.LanguageVersion); v != "" {
			report.LanguageVersion = v
		} else if v := strings.TrimSpace(m.FrameworkVersion); v != "" {
			report.LanguageVersion = v
		}
		if p := strings.TrimSpace(m.AppPort); p != "" {
			report.AppPort = p
		}
	}

	// 4) Опции — можно пробрасывать через ENV
	opts := pythonOpt{SonarHost: getenvDefault("SONAR_HOST_URL", ""), RegistryProject: getenvDefault("REGISTRY_PROJECT", "")}

	data := pythonTplData{Report: report, Opt: opts, Now: time.Now().Format(time.RFC3339), CI_COMMIT_REF_SLUG: "$CI_COMMIT_REF_SLUG"}

	// 5) Рендер
	tpl, err := template.New("python-ci").Option("missingkey=zero").Parse(string(raw))
	if err != nil {
		return fmt.Errorf("parse template: %w", err)
	}
	var buf bytes.Buffer
	if err := tpl.Execute(&buf, data); err != nil {
		return fmt.Errorf("execute template: %w", err)
	}

	// 6) Гарантировать, что docker job использует gentmp/Dockerfile
	yaml := buf.String()
	yaml = strings.ReplaceAll(yaml, "-f Dockerfile", "-f gentmp/Dockerfile")

	// 7) Сохранение и вывод
	outPath := filepath.Join(tmpDir, ".gitlab-ci.yml")
	if err := os.WriteFile(outPath, []byte(yaml), 0o644); err != nil {
		return fmt.Errorf("write .gitlab-ci.yml: %w", err)
	}

	fmt.Println("----- .gitlab-ci.yml -----")
	fmt.Print(yaml)
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
