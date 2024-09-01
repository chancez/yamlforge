/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"

	"github.com/chancez/yamlforge/pkg/yamlforge"
	"github.com/spf13/cobra"
)

type GenerateFlags struct {
	vars map[string]string
}

var genFlags GenerateFlags

// generateCmd represents the generate command
var generateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generate YAML",
	RunE: func(cmd *cobra.Command, args []string) error {
		forgeFile := "forge.yaml"
		if len(args) == 1 {
			forgeFile = args[0]
		}
		vars := make(map[string][]byte)
		for varName, varVal := range genFlags.vars {
			vars[varName] = []byte(varVal)
		}
		result, err := yamlforge.Generate(cmd.Context(), forgeFile, vars)
		if err != nil {
			return err
		}
		fmt.Println(string(result))
		return nil
	},
}

func init() {
	generateCmd.Flags().StringToStringVar(&genFlags.vars, "vars", nil, "Provide vars to the pipeline")
	rootCmd.AddCommand(generateCmd)
}
