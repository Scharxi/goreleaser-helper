package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "goreleaser-helper",
	Short: "A helper tool for managing Go releases",
	Long: `A helper tool for managing Go releases that provides a simplified interface
for creating and managing releases with proper versioning and changelog management.`,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	// Global flags can be added here
	rootCmd.PersistentFlags().StringP("config", "c", "", "Path to configuration file")
}
