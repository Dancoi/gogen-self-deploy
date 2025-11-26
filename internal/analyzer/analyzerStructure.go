package analyzer

import (
	"fmt"
)

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

type ProjectModule struct {

	Name string `json:"name"`

	ModulePath string `json:"module_path"`

	Language        Language `json:"language"`
	LanguageVersion string   `json:"language_version"`

	BuildTool BuildTool `json:"build_tool"`

	Framework        string `json:"framework"`
	FrameworkVersion string `json:"framework_version"`

	Dependencies []string `json:"dependencies"`

	BuildCommand string `json:"build_command"`

	TestCommand string `json:"test_command"`

	DockerfilePath string `json:"dockerfile_path"`

	BuilderImage string `json:"builder_image"`
	RuntimeImage string `json:"runtime_image"`

	ArtifactPath string `json:"artifact_path"`

	AppPort string `json:"app_port"`
}


type ProjectAnalysisResult struct {
	RepositoryName string `json:"repository_name"`


	Modules []*ProjectModule `json:"modules"`


	PipelineStrategy PipelineStrategy `json:"pipeline_strategy"`

	MainFramework        string `json:"main_framework"`
	MainFrameworkVersion string `json:"main_framework_version"`
}

func (par *ProjectAnalysisResult) PrintSummary() {
	fmt.Println("Project Analysis Summary:")
	fmt.Printf("Repository Name: %s\n", par.RepositoryName)
	fmt.Printf("Pipeline Strategy: %s\n", par.PipelineStrategy)
	fmt.Printf("Number of Modules: %d\n", len(par.Modules))
	for _, module := range par.Modules {
		fmt.Printf("- Module Name: %s, Language: %s, Build Tool: %s, Framework: %s\n",
			module.Name, module.Language, module.BuildTool, module.Framework)
	}
}
