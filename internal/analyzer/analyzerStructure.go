package analyzer

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
