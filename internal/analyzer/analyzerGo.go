package analyzer

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/Dancoi/gogen-self-deploy/internal/dto"
)

func AnalyzeGoModule(dto dto.RepoDTO, result *ProjectAnalysisResult) {
	root := dto.OutputDir 
	// root := "./tempfile"
	// root := filepath.Join(dto.OutputDir, dto.RepoName)
	fmt.Println("Analyzing Go modules in:", root)

	entries, err := os.ReadDir(root)
	if err != nil {
		log.Fatal(err)
	}

	result.RepositoryName = dto.RepoName
	

	for _, entry := range entries {
		if entry.IsDir() {
			path := filepath.Join(root, entry.Name(), "go.mod")

			if _, err := os.Stat(path); err != nil {
				continue 
			}

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
				BuilderImage: "golang:alpine", 
				RuntimeImage: "alpine:latest",
				ArtifactPath: "./app",
				AppPort:      "8080", 
			}

			reVersion := regexp.MustCompile(`go\s+([0-9]+\.[0-9]+)`)
			matchVersion := reVersion.FindStringSubmatch(strContent)
			if len(matchVersion) > 1 {
				module.LanguageVersion = matchVersion[1]
				module.BuilderImage = "golang:" + matchVersion[1] + "-alpine"
			}

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

			dockerPath := filepath.Join(filepath.Dir(path), "Dockerfile")
			if _, err := os.Stat(dockerPath); err == nil {
				module.DockerfilePath = dockerPath
			}

			result.Modules = append(result.Modules, module)
		}
	}
}
