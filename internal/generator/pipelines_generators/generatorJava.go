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

// GenerateJavaPipeline генерирует GitLab CI для Java (Maven/Gradle) + docker job.
func GenerateJavaPipeline(repoName string, repoRoot string, analysis *analyzer.ProjectAnalysisResult) error {
	_, err := dockerfiles_generators.GenerateJavaDockerfile(repoRoot, analysis)
	if err != nil {
		return fmt.Errorf("generate java dockerfile: %w", err)
	}

	tmpDir := "gentmp"
	if err := os.MkdirAll(tmpDir, 0o755); err != nil {
		return fmt.Errorf("mkdir gentmp: %w", err)
	}

	// Определяем build tool
	buildTool := "maven"
	javaVersion := "17"
	appName := repoName
	if analysis != nil {
		for _, m := range analysis.Modules {
			if m.Language == analyzer.LanguageJava {
				bt := strings.ToLower(string(m.BuildTool))
				if strings.Contains(bt, "gradle") {
					buildTool = "gradle"
				}
				if v := strings.TrimSpace(m.LanguageVersion); v != "" {
					javaVersion = sanitizeJavaVersion(v)
				}
				if n := strings.TrimSpace(m.Name); n != "" {
					appName = n
				}
				break
			}
		}
	}
	appName = sanitizeNameJava(appName)

	var tplPath string
	if buildTool == "gradle" {
		tplPath = filepath.Join("templates", "gitlab", "pipelines", "java_gradle.gitlab-ci.yml.tmpl")
	} else {
		tplPath = filepath.Join("templates", "gitlab", "pipelines", "java_maven.gitlab-ci.yml.tmpl")
	}
	data, err := os.ReadFile(tplPath)
	if err != nil {
		return fmt.Errorf("read java pipeline template: %w", err)
	}
	yaml := string(data)

	yaml = renderWithDefaultsJava(yaml, map[string]string{
		"JAVA_VERSION": javaVersion,
		"APP_NAME":     sanitizeNameJava(appName),
		"JAR_PATH":     chooseJarPath(buildTool),
	})
	if !strings.Contains(yaml, "gentmp/Dockerfile") {
		yaml = strings.ReplaceAll(yaml, "Dockerfile", "gentmp/Dockerfile")
	}

	outPath := filepath.Join(tmpDir, ".gitlab-ci.yml")
	if err := os.WriteFile(outPath, []byte(yaml), 0o644); err != nil {
		return fmt.Errorf("write java pipeline: %w", err)
	}
	fmt.Println("----- .gitlab-ci.yml (java) -----")
	fmt.Println(yaml)
	fmt.Println("----- end -----")
	fmt.Println("Saved to:", outPath)
	return nil
}

func chooseJarPath(tool string) string {
	if tool == "gradle" {
		return "build/libs/*.jar"
	}
	return "target/*.jar"
}

func sanitizeJavaVersion(v string) string {
	v = strings.TrimSpace(v)
	re := regexp.MustCompile(`\d+`)
	m := re.FindString(v)
	if m == "" {
		return "17"
	}
	return m
}

func sanitizeNameJava(s string) string {
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

func renderWithDefaultsJava(tpl string, vars map[string]string) string {
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
