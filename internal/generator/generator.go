package generator

import (
	"encoding/json"
	"fmt"
	"github.com/Dancoi/gogen-self-deploy/internal/analyzer"
	"github.com/Dancoi/gogen-self-deploy/internal/dto"
	"github.com/Dancoi/gogen-self-deploy/internal/fetcher"
	"io"
	"os"
	"path/filepath"
	"strings"
	"text/template"
	"time"

	"github.com/pkg/errors"
)

// PipelineTemplateData данные для шаблонов GitLab CI.
type PipelineTemplateData struct {
	Report dto.AnalyzeDTO
	Opt    dto.GenerationOptions
	Now    string
}

// Run полный цикл: clone -> analyze -> emit JSON -> generate pipeline.
func Run(repoURL, outputDir string, opt dto.GenerationOptions) error {
	if repoURL == "" {
		return errors.New("repoURL empty")
	}
	if outputDir == "" {
		outputDir = "./out"
	}
	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		return errors.Wrap(err, "mkdir outputDir")
	}

	// 1) Clone
	if err := fetcher.CloneRepo(repoURL, outputDir); err != nil {
		return errors.Wrap(err, "clone repo")
	}
	repoName := repoBaseName(repoURL)
	repoLocalPath := filepath.Join(outputDir, repoName)

	// 2) Подготовка временной структуры для текущего Go-анализатора (он читает ./tempfile)
	if err := prepareTempfileMirror(repoLocalPath); err != nil {
		return errors.Wrap(err, "prepare tempfile mirror")
	}

	// 3) Анализ
	res := &analyzer.ProjectAnalysisResult{
		RepositoryName: repoName,
	}
	analyzer.AnalyzRepo(res)

	// 4) Конверсия в DTO + JSON
	report := convertAnalysisToDTO(res, opt)
	if err := writeJSON(filepath.Join(outputDir, "analysis.json"), report); err != nil {
		return errors.Wrap(err, "write analysis.json")
	}

	// 5) Генерация пайплайна
	if err := generateGitlabPipeline(outputDir, report, opt); err != nil {
		return errors.Wrap(err, "generate pipeline")
	}

	fmt.Println("OK: analysis + pipeline готовы в:", outputDir)
	return nil
}

// repoBaseName выдергивает имя из URL.
func repoBaseName(url string) string {
	url = strings.TrimSpace(url)
	url = strings.TrimSuffix(url, "/")
	if i := strings.LastIndexAny(url, "/:"); i != -1 {
		base := url[i+1:]
		return strings.TrimSuffix(base, ".git")
	}
	return strings.TrimSuffix(url, ".git")
}

// prepareTempfileMirror создаёт ./tempfile и копирует туда директории с go.mod (для текущей реализации AnalyzeGoModule).
func prepareTempfileMirror(repoRoot string) error {
	tmpRoot := "./tempfile"
	_ = os.RemoveAll(tmpRoot)
	if err := os.MkdirAll(tmpRoot, 0o755); err != nil {
		return err
	}
	entries, err := os.ReadDir(repoRoot)
	if err != nil {
		return err
	}
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		dirPath := filepath.Join(repoRoot, e.Name())
		goMod := filepath.Join(dirPath, "go.mod")
		if _, err := os.Stat(goMod); err == nil {
			target := filepath.Join(tmpRoot, e.Name())
			if err := copyDir(dirPath, target); err != nil {
				return errors.Wrap(err, "copyDir "+dirPath)
			}
		}
	}
	return nil
}

// copyDir простое рекурсивное копирование.
func copyDir(src, dst string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		rel, _ := filepath.Rel(src, path)
		target := filepath.Join(dst, rel)
		if info.IsDir() {
			return os.MkdirAll(target, 0o755)
		}
		in, err := os.Open(path)
		if err != nil {
			return err
		}
		defer in.Close()
		out, err := os.Create(target)
		if err != nil {
			return err
		}
		defer out.Close()
		if _, err := io.Copy(out, in); err != nil {
			return err
		}
		return nil
	})
}

// convertAnalysisToDTO маппинг из внутреннего результата в контракт.
func convertAnalysisToDTO(res *analyzer.ProjectAnalysisResult, opt dto.GenerationOptions) dto.AnalyzeDTO {
	report := dto.AnalyzeDTO{}
	if len(res.Modules) == 0 {
		report.Language = dto.LangGo // fallback
		report.Docker = dto.DockerMeta{}
		return report
	}

	// Берём первый модуль как основной.
	m := res.Modules[0]
	report.Language = string(m.Language)
	report.Framework = m.Framework
	report.BuildTool = string(m.BuildTool)
	report.LanguageVersion = m.LanguageVersion
	report.AppPort = parsePort(m.AppPort)
	report.DetectedFiles = collectDetectedFiles(res)

	// Docker meta
	report.Docker = dto.DockerMeta{
		DockerfileDetected: m.DockerfilePath != "",
		PreferredBase:      dto.BaseAlpine,
		ExposedPort:        report.AppPort,
		ImageName:          opt.RegistryProject,
	}

	// Node/Python частные поля можно расширять по анализу (заглушки).
	switch report.CanonicalLanguage() {
	case dto.LangPython:
		report.Python = &dto.PythonMeta{
			RuntimeVersion: report.LanguageVersion,
			Framework:      m.Framework,
		}
	case dto.LangNode:
		report.Node = &dto.NodeMeta{
			NodeVersion: report.LanguageVersion,
			Framework:   m.Framework,
		}
	}

	return report
}

func parsePort(s string) int {
	if s == "" {
		return 0
	}
	var p int
	fmt.Sscanf(s, "%d", &p)
	return p
}

func collectDetectedFiles(res *analyzer.ProjectAnalysisResult) []string {
	set := make(map[string]struct{})
	for _, m := range res.Modules {
		if m.DockerfilePath != "" {
			set["Dockerfile"] = struct{}{}
		}
		if m.ModulePath != "" {
			set[filepath.Base(m.ModulePath)] = struct{}{}
		}
	}
	out := make([]string, 0, len(set))
	for k := range set {
		out = append(out, k)
	}
	return out
}

// writeJSON сохраняет DTO.
func writeJSON(path string, v any) error {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}

// generateGitlabPipeline рендерит .gitlab-ci.yml из соответствующего шаблона.
func generateGitlabPipeline(outputDir string, report dto.AnalyzeDTO, opt dto.GenerationOptions) error {
	tplPath, err := chooseTemplate(report)
	if err != nil {
		return err
	}
	tpl, err := template.New(filepath.Base(tplPath)).ParseFiles(tplPath)
	if err != nil {
		return errors.Wrap(err, "parse template")
	}
	data := PipelineTemplateData{
		Report: report,
		Opt:    opt,
		Now:    time.Now().Format(time.RFC3339),
	}
	outFile := filepath.Join(outputDir, ".gitlab-ci.yml")
	f, err := os.Create(outFile)
	if err != nil {
		return err
	}
	defer f.Close()
	if err := tpl.Execute(f, data); err != nil {
		return errors.Wrap(err, "execute template")
	}
	return nil
}

func chooseTemplate(report dto.AnalyzeDTO) (string, error) {
	switch report.CanonicalLanguage() {
	case dto.LangPython:
		return "templates/gitlab/pipelines/python.gitlab-ci.yml.tmpl", nil
	case dto.LangNode:
		// Один шаблон на Node/TypeScript
		return "templates/gitlab/pipelines/typescript.gitlab-ci.yml.tmpl", nil
	default:
		return "", errors.Errorf("нет шаблона для языка: %s", report.Language)
	}
}
