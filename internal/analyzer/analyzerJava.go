package analyzer

import (
	"log"
	"os"
	"path/filepath"
	"strings"
)

// containsString проверяет, содержит ли срез строк указанное значение.
func containsString(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func AnalyzeJava(result *ProjectAnalysisResult, start string) {
	targetFiles := []string{"pom.xml", "build.gradle", "build.gradle.kts"}

	err := filepath.WalkDir(start, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Ограничиваем глубину, чтобы не лазить в тесты/примеры слишком глубоко
		rel, relErr := filepath.Rel(start, path)
		if relErr != nil {
			return relErr
		}
		depth := strings.Count(rel, string(os.PathSeparator))
		if depth > 2 {
			if d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		if !d.IsDir() && containsString(targetFiles, d.Name()) {
			content, err := os.ReadFile(path)
			if err != nil {
				log.Printf("Не удалось прочитать %s по пути %s: %v", d.Name(), path, err)
				return nil
			}
			contentStr := string(content)

			module := &ProjectModule{
				Name:       filepath.Base(filepath.Dir(path)),
				ModulePath: path,
				Language:   LanguageJava,
			}

			if d.Name() == "pom.xml" {
				module.BuildTool = BuildToolMaven
				module.BuildCommand = "mvn clean install"
				module.TestCommand = "mvn test"
				module.BuilderImage = "maven:3.8.5-jdk-17"
				module.RuntimeImage = "eclipse-temurin:17-jre-alpine"

				// Версия Java: ищем java.version или maven.compiler.source
				if strings.Contains(contentStr, "<java.version>") {
					module.LanguageVersion = extractBetween(contentStr, "<java.version>", "</java.version>")
				} else if strings.Contains(contentStr, "<maven.compiler.source>") {
					module.LanguageVersion = extractBetween(contentStr, "<maven.compiler.source>", "</maven.compiler.source>")
				}

				// Определение Spring Boot и версии (по parent spring-boot-starter-parent)
				if strings.Contains(contentStr, "spring-boot-starter") {
					module.Framework = "spring-boot"
					if strings.Contains(contentStr, "<artifactId>spring-boot-starter-parent</artifactId>") {
						module.FrameworkVersion = extractBetween(contentStr, "<version>", "</version>")
					}
				}
			} else {
				// Gradle / Gradle Kotlin
				module.BuildTool = BuildToolGradle
				module.BuildCommand = "gradle build"
				module.TestCommand = "gradle test"
				module.BuilderImage = "gradle:7.6-jdk17"
				module.RuntimeImage = "eclipse-temurin:17-jre-alpine"

				// Версия Java: ищем sourceCompatibility/targetCompatibility
				for _, line := range strings.Split(contentStr, "\n") {
					line = strings.TrimSpace(line)
					if strings.HasPrefix(line, "sourceCompatibility") {
						module.LanguageVersion = strings.Trim(strings.TrimPrefix(line, "sourceCompatibility"), " =\"'")
					} else if strings.HasPrefix(line, "targetCompatibility") && module.LanguageVersion == "" {
						module.LanguageVersion = strings.Trim(strings.TrimPrefix(line, "targetCompatibility"), " =\"'")
					}
				}

				// Spring Boot: плагин или зависимости org.springframework.boot
				if strings.Contains(contentStr, "org.springframework.boot") {
					module.Framework = "spring-boot"
					// простая попытка вытащить версию из dependency "spring-boot-starter"
					if idx := strings.Index(contentStr, "spring-boot-starter"); idx != -1 {
						// оставляем FrameworkVersion пустым или дорабатываем позже
					}
				}
			}

			module.ArtifactPath = "./build/libs"
			module.AppPort = "8080"

			result.Modules = append(result.Modules, module)
			return filepath.SkipAll // нашли главный Java‑модуль
		}
		return nil
	})
	if err != nil {
		log.Printf("Ошибка при анализе Java модулей: %v", err)
	}
}

// extractBetween достает подстроку между двумя маркерами, если они найдены.
func extractBetween(s, start, end string) string {
	i := strings.Index(s, start)
	if i == -1 {
		return ""
	}
	i += len(start)
	j := strings.Index(s[i:], end)
	if j == -1 {
		return ""
	}
	return strings.TrimSpace(s[i : i+j])
}
