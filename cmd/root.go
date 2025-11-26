package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/Dancoi/gogen-self-deploy/internal/analyzer"
	"github.com/Dancoi/gogen-self-deploy/internal/dto"
	"github.com/Dancoi/gogen-self-deploy/internal/fetcher"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "gogen-self-deploy",
	Short: "Самостоятельный деплой",
	Long:  `gogen-self-deploy - это инструмент для самостоятельного деплоя приложений.`,

	Run: func(cmd *cobra.Command, args []string) {
		if len(args) < 2 {
			cmd.Help()
			return
		}
		repoURL := args[0]
		dir := args[1]

		DTO_Repo := dto.RepoDTO{
			RepoURL:   repoURL,
			OutputDir: dir,
			RepoName: fetcher.NameRepo(repoURL),
		}

		if err := fetcher.CloneRepo(DTO_Repo.RepoURL, DTO_Repo.OutputDir); err != nil {
			fmt.Println("Error cloning repository:", err)
			return
		}
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

		time.Sleep(2 * time.Second)
		if err := fetcher.DeleteRepo(DTO_Repo.RepoURL, DTO_Repo.OutputDir); err != nil {
			fmt.Println("Error deleting repository:", err)
			return
		}
	},
}
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	// Here you will define your flags and configuration settings.
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}


