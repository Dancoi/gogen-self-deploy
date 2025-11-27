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

// GenerateNodePipeline генерирует GitLab CI пайплайн для Node/TS проекта.
// 1) Генерация/копирование Dockerfile (gentmp/Dockerfile)
// 2) Рендер шаблона пайплайна templates/gitlab/pipelines/node.gitlab-ci.yml.tmpl
// 3) Подстановка плейсхолдеров ${NODE_VERSION}, ${APP_NAME}, ${BUILD_DIR}
// 4) Сохранение в gentmp/.gitlab-ci.yml и вывод
func GenerateNodePipeline(repoName string, repoRoot string, analysis *analyzer.ProjectAnalysisResult) error {
	_, err := dockerfiles_generators.GenerateNodeDockerfile(repoRoot, analysis)
	if err != nil {
		return fmt.Errorf("generate node dockerfile: %w", err)
	}

	// 1) Папка
	tmpDir := "gentmp"
	if err := os.MkdirAll(tmpDir, 0o755); err != nil {
		return fmt.Errorf("mkdir gentmp: %w", err)
	}

	// 2) Шаблон пайплайна
	tplPath := filepath.Join("templates", "gitlab", "pipelines", "node.gitlab-ci.yml.tmpl")
	data, err := os.ReadFile(tplPath)
	if err != nil {
		return fmt.Errorf("read node pipeline template: %w", err)
	}
	yaml := string(data)

	// 3) Из анализа
	nodeVersion := "20"
	appName := repoName
	buildDir := "dist"
	if analysis != nil && len(analysis.Modules) > 0 {
		for _, m := range analysis.Modules {
			if m.Language == analyzer.LanguageJavaScript || m.Language == analyzer.LanguageTypeScript {
				if v := strings.TrimSpace(m.LanguageVersion); v != "" {
					nodeVersion = v
				}
				break
			}
		}
	}
	appName = sanitizeName(appName)

	// 4) Подстановка
	yaml = renderWithDefaultsNode(yaml, map[string]string{
		"NODE_VERSION": nodeVersion,
		"APP_NAME":     appName,
		"BUILD_DIR":    buildDir,
	})

	// 5) Убедиться, что docker job использует наш Dockerfile путь (gentmp/Dockerfile)
	if !strings.Contains(yaml, "gentmp/Dockerfile") {
		// Простая замена, если вдруг шаблон другой.
		yaml = strings.ReplaceAll(yaml, "Dockerfile", "gentmp/Dockerfile")
	}

	// 6) Сохранение
	outPath := filepath.Join(tmpDir, ".gitlab-ci.yml")
	if err := os.WriteFile(outPath, []byte(yaml), 0o644); err != nil {
		return fmt.Errorf("write node pipeline: %w", err)
	}

	// 7) Вывод
	fmt.Println("----- .gitlab-ci.yml (node) -----")
	fmt.Println(yaml)
	fmt.Println("----- end -----")
	fmt.Println("Saved to:", outPath)
	return nil
}

func sanitizeName(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	if s == "" {
		return "app"
	}
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

func renderWithDefaultsNode(tpl string, vars map[string]string) string {
	reDef := regexp.MustCompile(`\$\{([A-Z0-9_]+):-[^}]*}`)
	tpl = reDef.ReplaceAllStringFunc(tpl, func(m string) string {
		key := reDef.FindStringSubmatch(m)[1]
		if v, ok := vars[key]; ok && v != "" {
			return v
		}
		parts := strings.SplitN(strings.TrimSuffix(strings.TrimPrefix(m, "${"), "}"), ":-", 2)
		if len(parts) == 2 {
			return parts[1]
		}
		return ""
	})
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
