package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "ai-report",
	Short: "Generate daily AI news reports",
}

var generateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generate the daily AI news report",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("generate not yet implemented")
		return nil
	},
}

var sourcesCmd = &cobra.Command{
	Use:   "sources",
	Short: "List configured news sources",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("sources not yet implemented")
		return nil
	},
}

func init() {
	generateCmd.Flags().String("date", "", "Generate report for a specific date (YYYY-MM-DD)")
	generateCmd.Flags().Bool("dry-run", false, "Show report without writing to file")
	rootCmd.AddCommand(generateCmd)
	rootCmd.AddCommand(sourcesCmd)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
