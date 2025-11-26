package analyzer

import (
	"encoding/json"
	"fmt"
)

// --- Enums ---

type Language string

const (
	LanguageGo         Language = "go"
	LanguagePython     Language = "python"
	LanguageJava       Language = "java"
	LanguageJavaScript Language = "javascript"
	LanguageTypeScript Language = "typescript"
	LanguageKotlin     Language = "kotlin"
	LanguageUnknown    Language = "unknown"
)

type BuildTool string

const (
	BuildToolMaven     BuildTool = "maven"
	BuildToolGradle    BuildTool = "gradle"
	BuildToolNpm       BuildTool = "npm"
	BuildToolYarn      BuildTool = "yarn"
	BuildToolPnpm      BuildTool = "pnpm"
	BuildToolPip       BuildTool = "pip"
	BuildToolPipenv    BuildTool = "pipenv"
	BuildToolPoetry    BuildTool = "poetry"
	BuildToolGoModules BuildTool = "go-modules"
	BuildToolUnknown   BuildTool = "unknown"
)

type PipelineStrategy string

const (
	PipelineStrategyMonorepo   PipelineStrategy = "monorepo"
	PipelineStrategyStandalone PipelineStrategy = "standalone"
)

// --- Structs ---

type ProjectModule struct {
	Name             string    `json:"name"`
	ModulePath       string    `json:"module_path"`
	Language         Language  `json:"language"`
	LanguageVersion  string    `json:"language_version"`
	BuildTool        BuildTool `json:"build_tool"`
	Framework        string    `json:"framework"`
	FrameworkVersion string    `json:"framework_version"`
	Dependencies     []string  `json:"dependencies"`
	BuildCommand     string    `json:"build_command"`
	TestCommand      string    `json:"test_command"`
	DockerfilePath   string    `json:"dockerfile_path"`
	BuilderImage     string    `json:"builder_image"`
	RuntimeImage     string    `json:"runtime_image"`
	ArtifactPath     string    `json:"artifact_path"`
	AppPort          string    `json:"app_port"`
}

type ProjectAnalysisResult struct {
	RepositoryName       string             `json:"repository_name"`
	Languages            map[string]float64 `json:"languages_percent"` // Статистика для "20 баллов"
	Infrastructure       []string           `json:"infrastructure"`    // Docker, K8s
	Modules              []*ProjectModule   `json:"modules"`
	PipelineStrategy     PipelineStrategy   `json:"pipeline_strategy"`
	MainFramework        string             `json:"main_framework"`
	MainFrameworkVersion string             `json:"main_framework_version"`
}

func (par *ProjectAnalysisResult) PrintSummary() {
	// Используем стандартный пакет encoding/json
	b, err := json.MarshalIndent(par, "", "  ")
	if err != nil {
		fmt.Println("Error marshaling result:", err)
		return
	}
	fmt.Println(string(b))
}

// shouldSkipDir - централизованная проверка игнорируемых папок
func shouldSkipDir(name string) bool {
	return name == ".git" || name == ".idea" || name == ".vscode" ||
		name == "node_modules" || name == "vendor" ||
		name == "dist" || name == "build" || name == "target" ||
		name == "__pycache__" || name == ".github"
}
