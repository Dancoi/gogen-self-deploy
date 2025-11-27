package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/Dancoi/gogen-self-deploy/internal/analyzer"
	"github.com/Dancoi/gogen-self-deploy/internal/dto"
	"github.com/Dancoi/gogen-self-deploy/internal/fetcher"
	"github.com/Dancoi/gogen-self-deploy/internal/generator/pipelines_generators"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "gogen-self-deploy",
	Short: "Самостоятельный деплой",
	Long:  `gogen-self-deploy - это инструмент для самостоятельного деплоя приложений.`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) < 2 {
			_ = cmd.Help()
			return
		}
		repoURL := args[0]
		dir := args[1]

		DTO_Repo := dto.RepoDTO{
			RepoURL:   repoURL,
			OutputDir: dir,
			RepoName:  fetcher.NameRepo(repoURL),
		}

		if err := fetcher.CloneRepo(DTO_Repo.RepoURL, DTO_Repo.OutputDir); err != nil {
			fmt.Println("Error cloning repository:", err)
			return
		}
		repoRoot := filepath.Join(DTO_Repo.OutputDir, DTO_Repo.RepoName)
		fmt.Println("Repository cloned successfully to", DTO_Repo.OutputDir)

		var analyzerRep *analyzer.ProjectAnalysisResult
		analyzerRep, err := analyzer.AnalyzRepo(DTO_Repo)
		if err != nil {
			fmt.Println("Error analyzing repository:", err)
			return
		}
		fmt.Println("Repository analyzed successfully")

		if analyzerRep != nil {
			analyzerRep.PrintSummary()
		}

		// Языковой выбор: node -> python -> go
		switch {
		case hasLanguage(analyzerRep, analyzer.LanguageJavaScript), hasLanguage(analyzerRep, analyzer.LanguageTypeScript):
			if err := pipelines_generators.GenerateNodePipeline(DTO_Repo.RepoName, repoRoot, analyzerRep); err != nil {
				fmt.Println("Error generating Node pipeline:", err)
			} else {
				fmt.Println("Node pipeline generated and printed")
			}
		case hasLanguage(analyzerRep, analyzer.LanguagePython):
			if err := pipelines_generators.GeneratePythonPipeline(analyzerRep); err != nil {
				fmt.Println("Error generating Python pipeline:", err)
			} else {
				fmt.Println("Python pipeline generated and printed")
			}
		case hasLanguage(analyzerRep, analyzer.LanguageGo):
			if err := pipelines_generators.GenerateGoPipeline(DTO_Repo.RepoName, repoRoot, analyzerRep); err != nil {
				fmt.Println("Error generating Go pipeline:", err)
			} else {
				fmt.Println("Go pipeline generated and printed")
			}
		default:
			fmt.Println("No supported languages detected for pipeline generation")
		}

		time.Sleep(2 * time.Second)
		if err := fetcher.DeleteRepo(DTO_Repo.RepoURL, DTO_Repo.OutputDir); err != nil {
			fmt.Println("Error deleting repository:", err)
			return
		}
	},
}

func hasLanguage(analysis *analyzer.ProjectAnalysisResult, lang analyzer.Language) bool {
	if analysis == nil {
		return false
	}
	for _, m := range analysis.Modules {
		if m.Language == lang {
			return true
		}
	}
	// fallback по глобальной статистике
	switch lang {
	case analyzer.LanguageJavaScript:
		if p, ok := analysis.Languages["JavaScript"]; ok && p > 0 {
			return true
		}
	case analyzer.LanguageTypeScript:
		if t, ok := analysis.Languages["TypeScript"]; ok && t > 0 {
			return true
		}
	case analyzer.LanguagePython:
		if p, ok := analysis.Languages["Python"]; ok && p > 0 {
			return true
		}
	case analyzer.LanguageGo:
		if g, ok := analysis.Languages["Go"]; ok && g > 0 {
			return true
		}
	}
	return false
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
