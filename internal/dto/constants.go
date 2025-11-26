package dto

// MVP языки (базовые)
const (
	LangGo         = "go"
	LangNode       = "node"       // покрывает JS/TS
	LangTypeScript = "typescript" // алиас к "node"
	LangPython     = "python"
	LangJava       = "java"
)

// Менеджеры пакетов / инструменты сборки
const (
	PmNpm    = "npm"
	PmYarn   = "yarn"
	PmPnpm   = "pnpm"
	PyPip    = "pip"
	PyPoetry = "poetry"
	PyPipenv = "pipenv"
)

// Варианты базовых образов
const (
	BaseAlpine     = "alpine"
	BaseSlim       = "slim"
	BaseDistroless = "distroless"
	BaseScratch    = "scratch"
)
