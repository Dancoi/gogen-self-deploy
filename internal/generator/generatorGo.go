package generator

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"text/template"

	"github.com/Dancoi/gogen-self-deploy/internal/analyzer"
)

// GenerateGoPipeline рендерит GitLab CI из go-шаблона, создаёт 'gentmp',
// генерирует мультистейдж Dockerfile (если в репо нет) и сохраняет 'gentmp/.gitlab-ci.yml'.
// repoRoot — путь до локально клонированного репозитория.
func GenerateGoPipeline(repoName string, repoRoot string, analysis *analyzer.ProjectAnalysisResult) error {
	// 1) Временная папка gentmp
	tmpDir := "gentmp"
	if err := os.MkdirAll(tmpDir, 0o755); err != nil {
		return fmt.Errorf("mkdir gentmp: %w", err)
	}

	// 2) Сгенерировать/скопировать Dockerfile
	dockerfilePath, err := ensureGoDockerfile(repoRoot, tmpDir, analysis)
	if err != nil {
		return fmt.Errorf("ensure go Dockerfile: %w", err)
	}

	// 3) Чтение шаблона пайплайна
	tplPath := filepath.Join("templates", "gitlab", "pipelines", "go.gitlab-ci.yml.tmpl")
	data, err := os.ReadFile(tplPath)
	if err != nil {
		return fmt.Errorf("read template: %w", err)
	}
	tpl := string(data)

	// 4) Достаём значения из анализа
	goVersion := "1.20"
	binaryName := "app"

	if analysis != nil && len(analysis.Modules) > 0 {
		m := analysis.Modules[0]
		versionCandidate := strings.TrimSpace(m.LanguageVersion)
		if versionCandidate == "" {
			versionCandidate = strings.TrimSpace(m.FrameworkVersion)
		}
		if versionCandidate != "" {
			goVersion = strings.TrimPrefix(versionCandidate, "go ")
		}
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
	}

	rendered := renderWithDefaults(tpl, map[string]string{
		"GOLANG_VERSION": goVersion,
		"BINARY_NAME":    binaryName,
	})

	rendered = appendGoDockerJob(rendered, dockerfilePath)

	outPath := filepath.Join(tmpDir, ".gitlab-ci.yml")
	if err := os.WriteFile(outPath, []byte(rendered), 0o644); err != nil {
		return fmt.Errorf("write .gitlab-ci.yml: %w", err)
	}

	fmt.Println("----- .gitlab-ci.yml -----")
	fmt.Println(rendered)
	fmt.Println("----- end -----")
	fmt.Println("Saved to:", outPath)
	return nil
}

func ensureGoDockerfile(repoRoot, tmpDir string, analysis *analyzer.ProjectAnalysisResult) (string, error) {
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
		fmt.Println("Saved Dockerfile to:", dst)
		fmt.Println("----- Dockerfile -----")
		fmt.Println(string(b))
		fmt.Println("----- end -----")
		return dst, nil
	}
	if analysis != nil && len(analysis.Modules) > 0 {
		m := analysis.Modules[0]
		if p := strings.TrimSpace(m.DockerfilePath); p != "" {
			abs := p
			if !filepath.IsAbs(abs) {
				abs = filepath.Join(repoRoot, p)
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
				fmt.Println("Saved Dockerfile to:", dst)
				fmt.Println("----- Dockerfile -----")
				fmt.Println(string(b))
				fmt.Println("----- end -----")
				return dst, nil
			}
		}
	}
	// Генерация из шаблона мультистейдж для Go
	// Пытаемся использовать существующий файл из templates; если не получится — упадём на встроенный fallback.
	tplPath := filepath.Join("templates", "dockerfiles", "go", "alpine", "Dockerfile_go_alpine.tmpl")
	raw, err := os.ReadFile(tplPath)
	if err != nil {
		// нет файла — используем встроенный мультистейдж
		return renderGoDockerfileFallback(tmpDir, analysis)
	}
	funcs := template.FuncMap{
		"default": func(val, def string) string {
			if strings.TrimSpace(val) == "" {
				return def
			}
			return val
		},
	}
	tpl, err := template.New("go-dockerfile").Funcs(funcs).Option("missingkey=zero").Parse(string(raw))
	if err != nil {
		return renderGoDockerfileFallback(tmpDir, analysis)
	}
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
	}
	var buf bytes.Buffer
	if err := tpl.Execute(&buf, data); err != nil {
		return renderGoDockerfileFallback(tmpDir, analysis)
	}
	dst := filepath.Join(tmpDir, "Dockerfile")
	if err := os.WriteFile(dst, buf.Bytes(), 0o644); err != nil {
		return "", err
	}
	fmt.Println("Saved Dockerfile to:", dst)
	fmt.Println("----- Dockerfile -----")
	fmt.Println(buf.String())
	fmt.Println("----- end -----")
	return dst, nil
}

func renderGoDockerfileFallback(tmpDir string, analysis *analyzer.ProjectAnalysisResult) (string, error) {
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
	content := "# Multistage Dockerfile (fallback) for Go\n" +
		"FROM golang:" + goVersion + "-alpine AS builder\n" +
		"WORKDIR /app\n" +
		"COPY . .\n" +
		"RUN CGO_ENABLED=0 GOOS=linux go build -ldflags=\"-s -w\" -o \"/out/" + binaryName + "\" ./...\n\n" +
		"FROM alpine:3.20 AS runtime\n" +
		"WORKDIR /app\n" +
		"COPY --from=builder /out/" + binaryName + " /app/" + binaryName + "\n"
	if appPort != "" {
		content += "EXPOSE " + appPort + "\n"
	}
	content += "ENTRYPOINT [\"/app/" + binaryName + "\"]\n"
	dst := filepath.Join(tmpDir, "Dockerfile")
	if err := os.WriteFile(dst, []byte(content), 0o644); err != nil {
		return "", err
	}
	fmt.Println("Saved Dockerfile to:", dst)
	fmt.Println("----- Dockerfile -----")
	fmt.Println(content)
	fmt.Println("----- end -----")
	return dst, nil
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


func renderWithDefaults(tpl string, vars map[string]string) string {
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

func appendGoDockerJob(yaml string, dockerfilePath string) string {
	if !strings.Contains(yaml, "stage: docker") && !strings.Contains(yaml, "- docker") {
		yaml += "\n\ndocker_build_push:\n  stage: docker\n  image: docker:24.0.7\n  services:\n    - name: docker:24.0.7-dind\n      command: [\"--tls=false\"]\n  variables:\n    DOCKER_DRIVER: overlay2\n  script:\n    - IMAGE=\"${CI_REGISTRY_IMAGE:-}\"\n    - TAG=\"${CI_COMMIT_SHORT_SHA:-local}\"\n    - if [ -z \"$IMAGE\" ]; then echo \"No image configured, set REGISTRY\"; exit 1; fi\n    - docker build -t \"$IMAGE:$TAG\" -f " + dockerfilePath + " .\n    - docker push \"$IMAGE:$TAG\"\n    - if [ \"$CI_COMMIT_BRANCH\" = \"main\" ] || [ \"$CI_COMMIT_BRANCH\" = \"master\" ]; then docker tag \"$IMAGE:$TAG\" \"$IMAGE:latest\"; docker push \"$IMAGE:latest\"; fi\n  only:\n    - branches\n"
	}
	return yaml
}
