package analyzer

import (
	"log"
	"os"
	"path/filepath"
	"strings"
)

func AnalyzeGoModule(result *ProjectAnalysisResult, start string) {
	targetFile := "go.mod"
	err := filepath.WalkDir(start, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() && d.Name() == targetFile {
			// Найден go.mod, анализируем его
			content, err := os.ReadFile(path)
			if err != nil {
				log.Printf("Не удалось прочитать go.mod по пути %s: %v", path, err)
				return nil
			}
			module := &ProjectModule{
				Name:       filepath.Base(filepath.Dir(path)),
				ModulePath: path,
			}
			lines := strings.Split(string(content), "\n")
			for _, line := range lines {
				line = strings.TrimSpace(line)
				if strings.HasPrefix(line, "module ") {
					module.Name = strings.TrimSpace(strings.TrimPrefix(line, "module "))
				} else if strings.HasPrefix(line, "go ") {
					module.FrameworkVersion = strings.TrimSpace(strings.TrimPrefix(line, "go "))
				} else if strings.HasPrefix(line, "require ") {
					parts := strings.Fields(strings.TrimPrefix(line, "require "))
					if len(parts) >= 2 {
						dep := parts[0] + " " + parts[1]
						module.Dependencies = append(module.Dependencies, dep)
					}
				}
			}
			module.Language = LanguageGo
			module.BuildTool = BuildToolGoModules
			module.BuildCommand = "go build ./..."
			module.TestCommand = "go test ./..."
			module.BuilderImage = "golang:1.20-alpine"
			module.RuntimeImage = "golang:1.20-alpine"
			module.ArtifactPath = "."
			module.AppPort = "8080"
			// Определение фреймворка
			goModStr := string(content)
			if strings.Contains(goModStr, "github.com/gin-gonic/gin") {
				module.Framework = "gin"
			} else if strings.Contains(goModStr, "github.com/labstack/echo") {
				module.Framework = "echo"
			} else if strings.Contains(goModStr, "github.com/gorilla/mux") {
				module.Framework = "mux"
			} else if strings.Contains(goModStr, "github.com/beego/beego") {
				module.Framework = "beego"
			}
			result.Modules = append(result.Modules, module)
			return filepath.SkipAll
		}
		return nil
	})
	if err != nil {
		log.Printf("Ошибка при обходе директорий для поиска go.mod: %v", err)
	}
}
