package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "notion-agent",
	Short: "Notion Agent — AI agent for your Notion workspace",
	Long: `Notion Agent is an AI-powered CLI that controls your Notion workspace.

Examples:
  notion-agent run "Crie uma database de tarefas com Status e Prazo"
  notion-agent run "Resuma todas as páginas da Sprint atual"
  notion-agent run "Adicione 5 tarefas de exemplo no Sprint Board"`,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
