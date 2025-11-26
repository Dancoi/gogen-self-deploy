package dto

// PythonMeta — детали Python-проекта для генератора.
type PythonMeta struct {
	DependencyManager string `json:"dependency_manager"` // "pip"|"poetry"|"pipenv"
	RuntimeVersion    string `json:"runtime_version"`    // 3.12 и т.п.
	HasPyproject      bool   `json:"has_pyproject"`
	HasRequirements   bool   `json:"has_requirements"`
	HasPipfile        bool   `json:"has_pipfile"`

	Framework    string `json:"framework"`     // "django"|"flask"|"fastapi"|"" (best-effort)
	TestCommand  string `json:"test_command"`  // pytest/unittest, если найдено
	LintCommand  string `json:"lint_command"`  // flake8/ruff/pylint
	StartCommand string `json:"start_command"` // run команда, если найдена

	WsgiModule string `json:"wsgi_module"` // для gunicorn ("app:app")
	AsgiModule string `json:"asgi_module"` // для uvicorn ("app:app")
}
