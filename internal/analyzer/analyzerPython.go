package analyzer

import (
	"bufio"
	"bytes"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

func AnalyzePythonModule(result *ProjectAnalysisResult, start string) {
	targetFiles := []string{"requirements.txt", "Pipfile", "pyproject.toml"}

	_ = filepath.WalkDir(start, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() && shouldSkipDir(d.Name()) {
			return filepath.SkipDir
		}

		rel, _ := filepath.Rel(start, path)
		if strings.Count(rel, string(os.PathSeparator)) > 3 {
			if d.IsDir() {
				return filepath.SkipDir
			}
		}

		if !d.IsDir() && containsString(targetFiles, d.Name()) {
			content, err := ioutil.ReadFile(path)
			if err != nil {
				return nil
			}

			module := &ProjectModule{
				Name:         filepath.Base(filepath.Dir(path)),
				ModulePath:   path,
				Language:     LanguagePython,
				BuilderImage: "python:3.11-slim",
				RuntimeImage: "python:3.11-slim",
				ArtifactPath: ".",
				AppPort:      "8000",
			}

			switch d.Name() {
			case "requirements.txt":
				module.BuildTool = BuildToolPip
				module.BuildCommand = "pip install -r requirements.txt"
				module.TestCommand = "pytest"
			case "Pipfile":
				module.BuildTool = BuildToolPipenv
				module.BuildCommand = "pipenv install"
				module.TestCommand = "pipenv run pytest"
			case "pyproject.toml":
				module.BuildTool = BuildToolPoetry
				module.BuildCommand = "poetry install"
				module.TestCommand = "poetry run pytest"
			}

			fw, ver := detectPythonFramework(content)
			if fw != "" {
				module.Framework = fw
				module.FrameworkVersion = ver
			}

			result.Modules = append(result.Modules, module)
			return filepath.SkipDir
		}
		return nil
	})
}

func detectPythonFramework(content []byte) (string, string) {
	scanner := bufio.NewScanner(bytes.NewReader(content))
	frameworks := []string{"django", "flask", "fastapi", "tornado", "pyramid", "starlette", "sanic"}

	for scanner.Scan() {
		line := strings.TrimSpace(strings.ToLower(scanner.Text()))
		if strings.HasPrefix(line, "#") || line == "" {
			continue
		}

		for _, fw := range frameworks {
			if line == fw || strings.HasPrefix(line, fw+"=") || strings.HasPrefix(line, fw+">") || strings.HasPrefix(line, fw+"<") {
				version := ""
				if strings.Contains(line, "==") {
					parts := strings.Split(line, "==")
					if len(parts) > 1 {
						version = strings.TrimSpace(parts[1])
					}
				}
				return capitalize(fw), version
			}
		}
	}
	return "", ""
}

func capitalize(s string) string {
	if len(s) == 0 {
		return ""
	}
	return strings.ToUpper(s[:1]) + s[1:]
}
