package analyzer

import (
	"encoding/xml"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

type PomProject struct {
	Parent       PomParent       `xml:"parent"`
	Properties   PomProperties   `xml:"properties"`
	Dependencies []PomDependency `xml:"dependencies>dependency"`
	Packaging    string          `xml:"packaging"`
}
type PomParent struct {
	GroupId    string `xml:"groupId"`
	ArtifactId string `xml:"artifactId"`
	Version    string `xml:"version"`
}
type PomDependency struct {
	GroupId    string `xml:"groupId"`
	ArtifactId string `xml:"artifactId"`
	Version    string `xml:"version"`
}
type PomProperties struct {
	JavaVersion         string `xml:"java.version"`
	MavenCompilerSource string `xml:"maven.compiler.source"`
}

func containsString(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func AnalyzeJavaModule(result *ProjectAnalysisResult, start string) {
	targetFiles := []string{"pom.xml", "build.gradle", "build.gradle.kts"}

	_ = filepath.WalkDir(start, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() && shouldSkipDir(d.Name()) {
			return filepath.SkipDir
		}

		// Ограничение глубины для оптимизации
		rel, _ := filepath.Rel(start, path)
		if strings.Count(rel, string(os.PathSeparator)) > 4 {
			if d.IsDir() {
				return filepath.SkipDir
			}
		}

		if !d.IsDir() && containsString(targetFiles, d.Name()) {
			module := &ProjectModule{
				Name:         filepath.Base(filepath.Dir(path)),
				ModulePath:   path,
				Language:     LanguageJava,
				ArtifactPath: "./target/*.jar",
				AppPort:      "8080",
			}

			content, _ := ioutil.ReadFile(path)

			if d.Name() == "pom.xml" {
				analyzeMaven(content, module)
			} else {
				analyzeGradle(string(content), module)
			}

			// --- ФИЛЬТР ШУМА ---
			// Добавляем модуль, только если это корневой модуль,
			// ИЛИ у него есть явный фреймворк (Quarkus/Spring),
			// ИЛИ это явно веб-приложение (war).
			// Иначе считаем это библиотекой внутри монорепо и пропускаем.
			isRoot := (filepath.Dir(path) == filepath.Clean(start))
			hasFramework := module.Framework != ""
			// Простая эвристика для WAR
			isWar := strings.Contains(string(content), "<packaging>war</packaging>")

			if isRoot || hasFramework || isWar {
				result.Modules = append(result.Modules, module)
			}

			return filepath.SkipDir
		}
		return nil
	})
}

func analyzeMaven(content []byte, module *ProjectModule) {
	module.BuildTool = BuildToolMaven
	module.BuildCommand = "mvn clean package -DskipTests"
	module.TestCommand = "mvn test"
	module.BuilderImage = "maven:3.9-eclipse-temurin-17"
	module.RuntimeImage = "eclipse-temurin:17-jre-alpine"

	var pom PomProject
	if err := xml.Unmarshal(content, &pom); err == nil {
		// Java Version
		if pom.Properties.JavaVersion != "" {
			module.LanguageVersion = pom.Properties.JavaVersion
		} else if pom.Properties.MavenCompilerSource != "" {
			module.LanguageVersion = pom.Properties.MavenCompilerSource
		} else {
			module.LanguageVersion = "17"
		}

		// Spring Boot Parent
		if pom.Parent.ArtifactId == "spring-boot-starter-parent" {
			module.Framework = "Spring Boot"
			module.FrameworkVersion = pom.Parent.Version
		}

		// Dependencies scan
		for _, dep := range pom.Dependencies {
			if module.Framework == "" && strings.Contains(dep.GroupId, "org.springframework.boot") {
				module.Framework = "Spring Boot"
			}
			if strings.Contains(dep.GroupId, "io.quarkus") {
				module.Framework = "Quarkus"
				if dep.Version != "" {
					module.FrameworkVersion = dep.Version
				}
			}
		}
	}
}

func analyzeGradle(content string, module *ProjectModule) {
	module.BuildTool = BuildToolGradle
	module.BuildCommand = "./gradlew build -x test"
	module.TestCommand = "./gradlew test"
	module.BuilderImage = "gradle:8.5-jdk17"
	module.RuntimeImage = "eclipse-temurin:17-jre-alpine"
	module.LanguageVersion = "17"

	if strings.Contains(content, "org.springframework.boot") {
		module.Framework = "Spring Boot"
	}
}
