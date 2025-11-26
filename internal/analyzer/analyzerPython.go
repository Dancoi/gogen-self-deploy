package analyzer

import (
	"log"
	"os"
	"path/filepath"
	"strings"
)

func AnalyzePython(result *ProjectAnalysisResult, start string) {
	targetFiles := []string{"requirements.txt", "Pipfile", "pyproject.toml"}

	err := filepath.WalkDir(start, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Ограничиваем глубину обхода (чтобы не лазить в глубокие тестовые/служебные директории)
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

			module := &ProjectModule{
				Name:       filepath.Base(filepath.Dir(path)),
				ModulePath: path,
				Language:   LanguagePython,
				// общие дефолты
				BuilderImage: "python:3.11-slim",
				RuntimeImage: "python:3.11-slim",
				ArtifactPath: ".",
				AppPort:      "8000",
			}

			switch d.Name() {
			case "requirements.txt":
				module.BuildTool = BuildToolPip
				module.BuildCommand = "pip install -r requirements.txt"
				module.TestCommand = "pytest tests/"
			case "Pipfile":
				module.BuildTool = BuildToolPipenv
				module.BuildCommand = "pipenv install"
				module.TestCommand = "pipenv run pytest tests/"
			case "pyproject.toml":
				module.BuildTool = BuildToolPoetry
				module.BuildCommand = "poetry install"
				module.TestCommand = "poetry run pytest tests/"
			}

			// Определение фреймворка (очень простая эвристика)
			pyProjStr := string(content)
			lower := strings.ToLower(pyProjStr)
			if strings.Contains(lower, "django") {
				module.Framework = "Django"
			} else if strings.Contains(lower, "flask") {
				module.Framework = "Flask"
			} else if strings.Contains(lower, "fastapi") {
				module.Framework = "FastAPI"
			}

			result.Modules = append(result.Modules, module)
			// считаем, что этого Python‑модуля достаточно
			return filepath.SkipAll
		}
		return nil
	})
	if err != nil {
		log.Printf("Ошибка при обходе директорий для Python: %v", err)
	}
}
