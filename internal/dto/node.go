package dto

// NodeMeta покрывает JS/TS.
type NodeMeta struct {
	PackageManager string `json:"package_manager"` // "npm"|"yarn"|"pnpm"
	NodeVersion    string `json:"node_version"`    // из engines.node
	HasTsconfig    bool   `json:"has_tsconfig"`

	Framework    string `json:"framework"` // "nextjs"|"nestjs"|"react"|"express"|"" (best-effort)
	BuildScript  bool   `json:"build_script"`
	TestScript   bool   `json:"test_script"`
	LintScript   bool   `json:"lint_script"`
	BuildDir     string `json:"build_dir"`     // "dist"|"build"|"" — для Docker/артефактов
	StartCommand string `json:"start_command"` // "node dist/index.js" и т.п.
}
