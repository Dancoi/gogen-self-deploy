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

// GeneratePythonPipeline рендерит GitLab CI из python-темплейта, создаёт 'gentmp',
// генерирует мультистейдж Dockerfile (если в репо нет) и сохраняет 'gentmp/.gitlab-ci.yml'.
// repoRoot — путь до локально клонированного репозитория.
func GeneratePythonPipeline(repoName string, repoRoot string, analysis *analyzer.ProjectAnalysisResult) error {
	// 1) Папка для вывода
	tmpDir := "gentmp"
	if err := os.MkdirAll(tmpDir, 0o755); err != nil {
		return fmt.Errorf("mkdir gentmp: %w", err)
	}

	// 2) Сгенерировать/скопировать Dockerfile
	dockerfilePath, err := ensurePythonDockerfile(repoRoot, tmpDir, analysis)
	if err != nil {
		return fmt.Errorf("ensure python Dockerfile: %w", err)
	}

	// 3) Чтение шаблона
	tplPath := filepath.Join("templates", "gitlab", "pipelines", "python.gitlab-ci.yml.tmpl")
	raw, err := os.ReadFile(tplPath)
	if err != nil {
		return fmt.Errorf("read template: %w", err)
	}

	// 4) Данные из анализа
	report := struct {
		LanguageVersion string
		AppPort         string
	}{
		LanguageVersion: "",
		AppPort:         "8000",
	}
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

	data := struct {
		Report             any
		Opt                struct{ SonarHost, RegistryProject string }
		Now                string
		CI_COMMIT_REF_SLUG string
	}{
		Report:             report,
		Opt:                struct{ SonarHost, RegistryProject string }{SonarHost: getenvDefault("SONAR_HOST_URL", ""), RegistryProject: getenvDefault("REGISTRY_PROJECT", "")},
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

	// 6) Добавить docker job
	buf.WriteString("\n\ndocker_build_push:\n  stage: docker\n  image: docker:24.0.7\n  services:\n    - name: docker:24.0.7-dind\n      command: [\"--tls=false\"]\n  variables:\n    DOCKER_DRIVER: overlay2\n  script:\n    - IMAGE=\"${CI_REGISTRY_IMAGE:-}\"\n    - TAG=\"${CI_COMMIT_SHORT_SHA:-local}\"\n    - if [ -z \"$IMAGE\" ]; then echo \"No image configured, set REGISTRY\"; exit 1; fi\n    - docker build -t \"$IMAGE:$TAG\" -f " + dockerfilePath + " .\n    - docker push \"$IMAGE:$TAG\"\n    - if [ \"$CI_COMMIT_BRANCH\" = \"main\" ] || [ \"$CI_COMMIT_BRANCH\" = \"master\" ]; then docker tag \"$IMAGE:$TAG\" \"$IMAGE:latest\"; docker push \"$IMAGE:latest\"; fi\n  only:\n    - branches\n")

	// 7) Сохранение и вывод
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

func ensurePythonDockerfile(repoRoot, tmpDir string, analysis *analyzer.ProjectAnalysisResult) (string, error) {
	// 1) Существующий Dockerfile
	existing := filepath.Join(repoRoot, "Dockerfile")
	if fi, err := os.Stat(existing); err == nil && !fi.IsDir() {
		dst := filepath.Join(tmpDir, "Dockerfile")
		b, err := os.ReadFile(existing)
		if err != nil {
			return "", err
		}
		if err := os.WriteFile(dst, b, 0o644); err != nil {
			return "", err
		}
		return dst, nil
	}
	// 2) Из анализа
	if analysis != nil {
		for _, m := range analysis.Modules {
			if m.Language == analyzer.LanguagePython && strings.TrimSpace(m.DockerfilePath) != "" {
				abs := m.DockerfilePath
				if !filepath.IsAbs(abs) {
					abs = filepath.Join(repoRoot, abs)
				}
				if fi, err := os.Stat(abs); err == nil && !fi.IsDir() {
					dst := filepath.Join(tmpDir, "Dockerfile")
					b, err := os.ReadFile(abs)
					if err != nil {
						return "", err
					}
					if err := os.WriteFile(dst, b, 0o644); err != nil {
						return "", err
					}
					return dst, nil
				}
			}
		}
	}
	// 3) Генерация из шаблона
	// Выбираем slim шаблон
	tplPath := filepath.Join("templates", "dockerfiles", "python", "slim", "Dockerfile_python_slim.tmpl")
	raw, err := os.ReadFile(tplPath)
	if err != nil {
		return "", fmt.Errorf("read python dockerfile template: %w", err)
	}
	tpl, err := template.New("python-dockerfile").Option("missingkey=zero").Parse(string(raw))
	if err != nil {
		return "", fmt.Errorf("parse python dockerfile template: %w", err)
	}
	pyVersion := "3.11"
	appPort := ""
	if m := firstPythonModule(analysis); m != nil {
		if v := strings.TrimSpace(m.LanguageVersion); v != "" {
			pyVersion = v
		}
		appPort = strings.TrimSpace(m.AppPort)
	}
	data := map[string]any{
		"PythonVersion":    pyVersion,
		"BaseImageBuilder": fmt.Sprintf("python:%s-slim", pyVersion),
		"BaseImageRuntime": fmt.Sprintf("python:%s-slim", pyVersion),
		"AppWorkdir":       "/app",
		"ExposePort":       appPort,
	}
	var buf bytes.Buffer
	if err := tpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("render python dockerfile: %w", err)
	}
	dst := filepath.Join(tmpDir, "Dockerfile")
	if err := os.WriteFile(dst, buf.Bytes(), 0o644); err != nil {
		return "", err
	}
	return dst, nil
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
