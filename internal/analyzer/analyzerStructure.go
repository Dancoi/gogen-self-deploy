package analyzer

import (
	"fmt"
)

// Языки программирования
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

// Инструменты сборки
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

// Стратергии пайплайна
type PipelineStrategy string

const (
	PipelineStrategyMonorepo   PipelineStrategy = "monorepo"
	PipelineStrategyStandalone PipelineStrategy = "standalone"
)

// Описание одного модуля проекта
type ProjectModule struct {
	//Имя модуля
	Name string `json:"name"`

	//Путь к модулю относительно корня проекта
	ModulePath string `json:"module_path"`

	//Язык программирования модуля и версия
	Language        Language `json:"language"`
	LanguageVersion string   `json:"language_version"`

	//Инструмент сборки модуля
	BuildTool BuildTool `json:"build_tool"`

	//Фреймворк модуля и версия
	Framework        string `json:"framework"`
	FrameworkVersion string `json:"framework_version"`

	//Зависимости модуля
	Dependencies []string `json:"dependencies"`

	//Команда сборки модуля
	BuildCommand string `json:"build_command"`

	//Команда теста модуля
	TestCommand string `json:"test_command"`

	//Путь к файлу Dockerfile модуля (если есть)
	DockerfilePath string `json:"dockerfile_path"`

	//Базовые образы для сборки и запуска модуля
	BuilderImage string `json:"builder_image"`
	RuntimeImage string `json:"runtime_image"`

	//Путь к артефакту сборки модуля
	ArtifactPath string `json:"artifact_path"`

	//Порт приложения модуля
	AppPort string `json:"app_port"`
}

// Результат анализа проекта
type ProjectAnalysisResult struct {
	//Имя репозитория проекта
	RepositoryName string `json:"repository_name"`

	//Список модулей проекта
	Modules []*ProjectModule `json:"modules"`

	//Стартегия пайплайна для проекта
	PipelineStrategy PipelineStrategy `json:"pipeline_strategy"`

	//Основной фреймворк и его версия (для одиночных проектов)
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
