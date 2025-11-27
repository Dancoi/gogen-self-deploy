package dockerfiles_generators

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/Dancoi/gogen-self-deploy/internal/analyzer"
	"github.com/Dancoi/gogen-self-deploy/internal/generator/util"
)

// GenerateJavaDockerfile генерирует или копирует мультистейдж Dockerfile для Java.
// Определяет инструмент сборки (Maven / Gradle) и берёт соответствующий шаблон:
// Maven: templates/dockerfiles/java/maven/distroless/Dockerfile_java_maven_multistage.tmpl
// Gradle: templates/dockerfiles/java/gradle/distroless/Dockerfile_java_gradle_multistage.tmpl
// Сохраняет результат в gentmp/Dockerfile.
func GenerateJavaDockerfile(repoRoot string, analysis *analyzer.ProjectAnalysisResult) (string, error) {
	tmpDir := "gentmp"
	if err := os.MkdirAll(tmpDir, 0o755); err != nil {
		return "", fmt.Errorf("mkdir gentmp: %w", err)
	}
	outPath := filepath.Join(tmpDir, "Dockerfile")

	// 1) Существующий Dockerfile
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

	// 2) Анализ модуля Java
	buildTool := "maven"
	javaVersion := "17"
	var module *analyzer.ProjectModule
	if analysis != nil {
		for _, m := range analysis.Modules {
			if m.Language == analyzer.LanguageJava {
				module = m
				break
			}
		}
	}
	if module != nil {
		bt := strings.ToLower(string(module.BuildTool))
		if strings.Contains(bt, "gradle") {
			buildTool = "gradle"
		}
		if v := strings.TrimSpace(module.LanguageVersion); v != "" {
			javaVersion = trimJavaVersion(v)
		}
	}

	var tplPath string
	if buildTool == "gradle" {
		tplPath = filepath.Join("templates", "dockerfiles", "java", "gradle", "distroless", "Dockerfile_java_gradle_multistage.tmpl")
	} else {
		tplPath = filepath.Join("templates", "dockerfiles", "java", "maven", "distroless", "Dockerfile_java_maven_multistage.tmpl")
	}

	raw, err := os.ReadFile(tplPath)
	if err != nil {
		return "", fmt.Errorf("read java dockerfile template: %w", err)
	}
	funcs := template.FuncMap{
		"default": func(def string, val string) string {
			if strings.TrimSpace(val) == "" {
				return def
			}
			return val
		},
	}
	tpl, err := template.New("java-dockerfile").Funcs(funcs).Option("missingkey=zero").Parse(string(raw))
	if err != nil {
		return "", fmt.Errorf("parse java dockerfile template: %w", err)
	}

	data := map[string]any{
		"JavaVersion":       javaVersion,
		"AppWorkdir":        "/app",
		"JarNamePattern":    "*.jar",
		"BuildTool":         buildTool,
		"RuntimeBaseImage":  fmt.Sprintf("gcr.io/distroless/java%s-debian12", majorJava(javaVersion)),
		"BaseImageBuilder":  fmt.Sprintf("eclipse-temurin:%s-jdk", majorJava(javaVersion)),
		"BaseImageRuntime":  fmt.Sprintf("eclipse-temurin:%s-jre", majorJava(javaVersion)),
		"MainClass":         "",
		"AdditionalRunArgs": []string{},
		"SkipTests":         "true",
	}
	var buf bytes.Buffer
	if err := tpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("render java dockerfile: %w", err)
	}
	if err := os.WriteFile(outPath, buf.Bytes(), 0o644); err != nil {
		return "", err
	}
	util.PrintDockerfile(outPath, buf.Bytes())
	return outPath, nil
}

func trimJavaVersion(v string) string {
	v = strings.TrimSpace(v)
	v = strings.TrimPrefix(strings.ToLower(v), "java")
	v = strings.TrimLeft(v, " :")
	if v == "" {
		return "17"
	}
	return v
}

func majorJava(v string) string {
	for i := 0; i < len(v); i++ {
		if v[i] < '0' || v[i] > '9' {
			return v[:i]
		}
	}
	return v
}
