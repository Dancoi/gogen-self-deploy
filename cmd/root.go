package cmd

import (
	"fmt"
	"os"

	"github.com/Dancoi/gogen-self-deploy/internal/fetcher"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "gogen-self-deploy",
	Short: "Самостоятельный деплой",
	Long: `gogen-self-deploy - это инструмент для самостоятельного деплоя приложений.`,
	
	Run: func(cmd *cobra.Command, args []string) { 
		if len(args) < 2 {
			cmd.Help()
			return
		}
		repoURL := args[0]
		dir := args[1]

		if err := fetcher.CloneRepo(repoURL, dir); err != nil {
			fmt.Println("Error cloning repository:", err)
			return
		}
		fmt.Println("Repository cloned successfully to", dir)
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


