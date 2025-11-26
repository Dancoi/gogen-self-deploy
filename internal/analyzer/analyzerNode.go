package analyzer

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

func AnalyzeNodeModule(result *ProjectAnalysisResult, start string) {
	targetFiles := []string{"package.json"}

	_ = filepath.WalkDir(start, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() && shouldSkipDir(d.Name()) {
			return filepath.SkipDir
		}

		if !d.IsDir() && containsString(targetFiles, d.Name()) {
			content, err := ioutil.ReadFile(path)
			if err != nil {
				return nil
			}

			type packageJSON struct {
				Name            string            `json:"name"`
				Scripts         map[string]string `json:"scripts"`
				Dependencies    map[string]string `json:"dependencies"`
				DevDependencies map[string]string `json:"devDependencies"`
				Engines         map[string]string `json:"engines"`
			}

			var pkg packageJSON
			if err := json.Unmarshal(content, &pkg); err != nil {
				return nil
			}

			module := &ProjectModule{
				Name:         filepath.Base(filepath.Dir(path)),
				ModulePath:   path,
				Language:     LanguageJavaScript,
				BuildTool:    BuildToolNpm,
				BuildCommand: "npm install && npm run build",
				TestCommand:  "npm test",
				BuilderImage: "node:18-alpine",
				RuntimeImage: "node:18-alpine",
				ArtifactPath: "dist",
				AppPort:      "3000",
			}

			if pkg.Name != "" {
				module.Name = pkg.Name
			}
			if _, ok := pkg.DevDependencies["typescript"]; ok {
				module.Language = LanguageTypeScript
			}

			if v, ok := pkg.Engines["node"]; ok {
				module.LanguageVersion = v
			} else {
				module.LanguageVersion = "18"
			}

			checkFrameworks := func(deps map[string]string) {
				for dep, ver := range deps {
					if strings.Contains(dep, "express") {
						module.Framework = "Express"
						module.FrameworkVersion = ver
					} else if strings.Contains(dep, "nestjs") {
						module.Framework = "NestJS"
						module.FrameworkVersion = ver
					} else if strings.Contains(dep, "next") {
						module.Framework = "Next.js"
						module.FrameworkVersion = ver
					} else if strings.Contains(dep, "react") && module.Framework == "" {
						module.Framework = "React"
						module.FrameworkVersion = ver
					} else if strings.Contains(dep, "vue") && module.Framework == "" {
						module.Framework = "Vue"
						module.FrameworkVersion = ver
					}
				}
			}
			checkFrameworks(pkg.Dependencies)
			result.Modules = append(result.Modules, module)
			return filepath.SkipDir
		}
		return nil
	})
}
