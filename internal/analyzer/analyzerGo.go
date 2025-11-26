package analyzer

import (
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

func AnalyzeGoModule(result *ProjectAnalysisResult) {
	//root := dto.RepoTDO.OutputDir
	root := "./tempfile"

	// Читаем содержимое корневой директории tempfile
	entries, err := os.ReadDir(root)
	if err != nil {
		log.Fatal(err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			// Формируем путь к предполагаемому главному go.mod (в корне репозитория)
			path := filepath.Join(root, entry.Name(), "go.mod")

			// Проверяем, существует ли файл
			if _, err := os.Stat(path); err != nil {
				continue // Если файла нет, пропускаем
			}

			// Поиск go.mod, анализируем его
			content, err := os.ReadFile(path)
			if err != nil {
				log.Printf("Не удалось прочитать go.mod по пути %s: %v", path, err)
				continue
			}
			strContent := string(content)

			module := &ProjectModule{
				Name:         filepath.Base(filepath.Dir(path)),
				ModulePath:   path,
				Language:     LanguageGo,
				BuildTool:    BuildToolGoModules,
				BuildCommand: "go build -o app ./...",
				TestCommand:  "go test ./...",
				Dependencies: []string{},
				BuilderImage: "golang:alpine", // Значение по умолчанию
				RuntimeImage: "alpine:latest",
				ArtifactPath: "./app",
				AppPort:      "8080", // Порт по умолчанию, сложно определить статически
			}

			// Парсинг версии Go
			reVersion := regexp.MustCompile(`go\s+([0-9]+\.[0-9]+)`)
			matchVersion := reVersion.FindStringSubmatch(strContent)
			if len(matchVersion) > 1 {
				module.LanguageVersion = matchVersion[1]
				// Обновляем образ сборки с учетом версии
				module.BuilderImage = "golang:" + matchVersion[1] + "-alpine"
			}

			// Парсинг зависимостей
			reDep := regexp.MustCompile(`\s+([a-zA-Z0-9.\-_/]+)\s+(v[0-9.]+)`)
			matches := reDep.FindAllStringSubmatch(strContent, -1)
			for _, m := range matches {
				if len(m) > 2 {
					module.Dependencies = append(module.Dependencies, m[1]+"@"+m[2])
					if strings.Contains(m[1], "gin-gonic/gin") {
						module.Framework = "gin"
						module.FrameworkVersion = m[2]
					} else if strings.Contains(m[1], "labstack/echo") {
						module.Framework = "echo"
						module.FrameworkVersion = m[2]
					} else if strings.Contains(m[1], "gofiber/fiber") {
						module.Framework = "fiber"
						module.FrameworkVersion = m[2]
					}
				}
			}

			// Проверка наличия Dockerfile
			dockerPath := filepath.Join(filepath.Dir(path), "Dockerfile")
			if _, err := os.Stat(dockerPath); err == nil {
				module.DockerfilePath = dockerPath
			}

			result.Modules = append(result.Modules, module)
		}
	}
}
