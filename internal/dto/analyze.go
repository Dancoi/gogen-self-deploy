package dto

// AnalyzeDTO — единый контракт между анализатором и генератором.
type AnalyzeDTO struct {
	// Общая часть
	Language        string   `json:"language"`         // см. constants.go
	Framework       string   `json:"framework"`        // общая наводка
	BuildTool       string   `json:"build_tool"`       // для java/go и т.п.
	LanguageVersion string   `json:"language_version"` // best-effort
	AppPort         int      `json:"app_port"`
	DetectedFiles   []string `json:"detected_files"`

	// Детали по языкам
	Python *PythonMeta `json:"python,omitempty"`
	Node   *NodeMeta   `json:"node,omitempty"`

	// Контейнеризация
	Docker DockerMeta `json:"docker"`
}

// CanonicalLanguage нормализует алиасы (например, "typescript" -> "node").
func (a AnalyzeDTO) CanonicalLanguage() string {
	switch a.Language {
	case LangTypeScript:
		return LangNode
	default:
		return a.Language
	}
}
