package cli

import (
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "forge",
	Short: "Forge is an AI-powered software engineering assistant",
}

// Execute ejecuta el comando raíz del CLI de Forge.
func Execute() error {
	return rootCmd.Execute()
}
