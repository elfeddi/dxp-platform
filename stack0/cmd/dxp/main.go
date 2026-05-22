package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "dxp",
	Short: "DxP — Engineering Platform as a Service",
	Long:  `DxP CLI — gérez votre plateforme d'ingénierie via dxp.yaml`,
}

var validateCmd = &cobra.Command{
	Use:   "validate [fichier]",
	Short: "Valide un fichier dxp.yaml",
	Args:  cobra.ExactArgs(1),
	RunE:  runValidate,
}

var applyCmd = &cobra.Command{
	Use:   "apply [fichier]",
	Short: "Applique un fichier dxp.yaml sur le cluster",
	Args:  cobra.ExactArgs(1),
	RunE:  runApply,
}

var generateCmd = &cobra.Command{
	Use:   "generate [prompt]",
	Short: "Génère un dxp.yaml depuis un prompt naturel",
	Args:  cobra.ExactArgs(1),
	RunE:  runGenerate,
}

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Affiche l'état des stacks DxP",
	RunE:  runStatus,
}

func main() {
	rootCmd.AddCommand(validateCmd)
	rootCmd.AddCommand(applyCmd)
	rootCmd.AddCommand(generateCmd)
	rootCmd.AddCommand(statusCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
