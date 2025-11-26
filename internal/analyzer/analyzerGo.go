package analyzer

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/mod/modfile"
)

func AnalyzeGoModule(result *ProjectAnalysisResult, start string) {
	targetFile := "go.mod"

	_ = filepath.WalkDir(start, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() && shouldSkipDir(d.Name()) {
			return filepath.SkipDir
		}

		if !d.IsDir() && d.Name() == targetFile {
			content, err := ioutil.ReadFile(path)
			if err != nil {
				return nil
			}

			f, err := modfile.Parse(path, content, nil)
			if err != nil {
				return nil
			}

			module := &ProjectModule{
				Name:            filepath.Base(filepath.Dir(path)),
				ModulePath:      path,
				Language:        LanguageGo,
				BuildTool:       BuildToolGoModules,
				BuildCommand:    "go build -o app ./...",
				TestCommand:     "go test ./...",
				ArtifactPath:    ".",
				AppPort:         "8080",
				LanguageVersion: "1.21",
			}

			if f.Module != nil {
				module.Name = f.Module.Mod.Path
			}
			if f.Go != nil {
				module.LanguageVersion = f.Go.Version
				module.BuilderImage = "golang:" + f.Go.Version + "-alpine"
			}

			for _, req := range f.Require {
				module.Dependencies = append(module.Dependencies, req.Mod.Path)
				if strings.Contains(req.Mod.Path, "gin-gonic/gin") {
					module.Framework = "Gin"
					module.FrameworkVersion = req.Mod.Version
				} else if strings.Contains(req.Mod.Path, "labstack/echo") {
					module.Framework = "Echo"
					module.FrameworkVersion = req.Mod.Version
				} else if strings.Contains(req.Mod.Path, "gofiber/fiber") {
					module.Framework = "Fiber"
					module.FrameworkVersion = req.Mod.Version
				} else if strings.Contains(req.Mod.Path, "gorilla/mux") {
					module.Framework = "Gorilla Mux"
					module.FrameworkVersion = req.Mod.Version
				}
			}

			result.Modules = append(result.Modules, module)
			return filepath.SkipDir
		}
		return nil
	})
}
