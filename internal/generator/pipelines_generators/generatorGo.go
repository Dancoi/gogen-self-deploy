package pipelines_generators

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/Dancoi/gogen-self-deploy/internal/analyzer"
	"github.com/Dancoi/gogen-self-deploy/internal/generator/dockerfiles_generators"
)

// GenerateGoPipeline рендерит GitLab CI из go-шаблона, создаёт 'gentmp',
// сохраняет 'gentmp/.gitlab-ci.yml', печатает его в консоль и добавляет docker-джобу,
// используя сгенерированный или существующий Dockerfile.
func GenerateGoPipeline(repoName string, repoRoot string, analysis *analyzer.ProjectAnalysisResult) error {
	// 0) Сначала сгенерируем/скопируем Dockerfile в gentmp
	dockerfilePath, err := dockerfiles_generators.GenerateGoDockerfile(repoRoot, analysis)
	if err != nil {
		return fmt.Errorf("generate go dockerfile: %w", err)
	}

	// 1) Временная папка gentmp
	tmpDir := "gentmp"
	if err := os.MkdirAll(tmpDir, 0o755); err != nil {
		return fmt.Errorf("mkdir gentmp: %w", err)
	}

	// 2) Чтение шаблона
	tplPath := filepath.Join("templates", "gitlab", "pipelines", "go.gitlab-ci.yml.tmpl")
	data, err := os.ReadFile(tplPath)
	if err != nil {
		return fmt.Errorf("read template: %w", err)
	}
	tpl := string(data)

	// 3) Достаём значения из анализа
	goVersion := "1.20"
	binaryName := "app"

	if analysis != nil && len(analysis.Modules) > 0 {
		m := analysis.Modules[0]
		// Версия: сначала LanguageVersion, если пусто - берём FrameworkVersion (в текущем анализаторе туда ложится версия Go)
		versionCandidate := strings.TrimSpace(m.LanguageVersion)
		if versionCandidate == "" {
			versionCandidate = strings.TrimSpace(m.FrameworkVersion)
		}
		if versionCandidate != "" {
			goVersion = strings.TrimPrefix(versionCandidate, "go ")
		}
		// Имя бинарника: последний сегмент из module name (m.Name) либо имя репозитория.
		rawName := strings.TrimSpace(m.Name)
		if rawName != "" {
			parts := strings.Split(rawName, "/")
			last := parts[len(parts)-1]
			if last != "" {
				binaryName = sanitizeBinaryName(last)
			}
		} else if rn := strings.TrimSpace(repoName); rn != "" {
			binaryName = sanitizeBinaryName(rn)
		}
	} else if strings.TrimSpace(repoName) != "" {
		binaryName = sanitizeBinaryName(repoName)
	}

	// 4) Подстановка плейсхолдеров ${VAR} и ${VAR:-default}
	rendered := renderWithDefaults(tpl, map[string]string{
		"GOLANG_VERSION": goVersion,
		"BINARY_NAME":    binaryName,
	})

	// 5) Добавляем docker-джобу, если её нет
	rendered = appendGoDockerJob(rendered, dockerfilePath)

	// 6) Сохранение в gentmp/.gitlab-ci.yml
	outPath := filepath.Join(tmpDir, ".gitlab-ci.yml")
	if err := os.WriteFile(outPath, []byte(rendered), 0o644); err != nil {
		return fmt.Errorf("write .gitlab-ci.yml: %w", err)
	}

	// 7) Вывод содержимого в консоль
	fmt.Println("----- .gitlab-ci.yml -----")
	fmt.Println(rendered)
	fmt.Println("----- end -----")
	fmt.Println("Saved to:", outPath)
	return nil
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

// renderWithDefaults заменяет ${KEY} и ${KEY:-default} на значения из vars.
func renderWithDefaults(tpl string, vars map[string]string) string {
	// ${KEY:-default}
	reDef := regexp.MustCompile(`\$\{([A-Z0-9_]+):-[^}]*}`)
	tpl = reDef.ReplaceAllStringFunc(tpl, func(m string) string {
		key := reDef.FindStringSubmatch(m)[1]
		if v, ok := vars[key]; ok && v != "" {
			return v
		}
		// Если нет значения - оставим дефолт (отрежем префикс до :- и вернём default часть)
		// Формат: ${KEY:-default}
		parts := strings.SplitN(strings.TrimSuffix(strings.TrimPrefix(m, "${"), "}"), ":-", 2)
		if len(parts) == 2 {
			return parts[1]
		}
		return ""
	})
	// ${KEY}
	reVar := regexp.MustCompile(`\$\{([A-Z0-9_]+)}`)
	tpl = reVar.ReplaceAllStringFunc(tpl, func(m string) string {
		key := reVar.FindStringSubmatch(m)[1]
		if v, ok := vars[key]; ok {
			return v
		}
		return ""
	})
	return tpl
}

// appendGoDockerJob добавляет docker stage+job, если их нет, используя указанный Dockerfile
func appendGoDockerJob(yaml string, dockerfilePath string) string {
	if !strings.Contains(yaml, "stage: docker") && !strings.Contains(yaml, "- docker") {
		yaml += "\n\ndocker_build_push:\n  stage: docker\n  image: docker:24.0.7\n  services:\n    - name: docker:24.0.7-dind\n      command: [\"--tls=false\"]\n  variables:\n    DOCKER_DRIVER: overlay2\n  script:\n    - IMAGE=\"${CI_REGISTRY_IMAGE:-}\"\n    - TAG=\"${CI_COMMIT_SHORT_SHA:-local}\"\n    - if [ -z \"$IMAGE\" ]; then echo \"No image configured, set REGISTRY\"; exit 1; fi\n    - docker build -t \"$IMAGE:$TAG\" -f " + dockerfilePath + " .\n    - docker push \"$IMAGE:$TAG\"\n    - if [ \"$CI_COMMIT_BRANCH\" = \"main\" ] || [ \"$CI_COMMIT_BRANCH\" = \"master\" ]; then docker tag \"$IMAGE:$TAG\" \"$IMAGE:latest\"; docker push \"$IMAGE:latest\"; fi\n  only:\n    - branches\n"
	}
	return yaml
}
