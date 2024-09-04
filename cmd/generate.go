/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"io"
	"path"

	"github.com/chancez/yamlforge/pkg/config"
	"github.com/chancez/yamlforge/pkg/generator"
	"github.com/chancez/yamlforge/pkg/reference"
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

		cfg, err := config.ParseFile(forgeFile)
		if err != nil {
			return fmt.Errorf("error parsing pipeline %s: %w", forgeFile, err)
		}

		refStore := reference.NewStore(vars)
		state := generator.NewPipeline(path.Dir(forgeFile), cfg.PipelineGenerator, refStore)
		result, err := state.Generate(cmd.Context())
		if err != nil {
			return err
		}

		_, err = io.WriteString(cmd.OutOrStdout(), string(result))
		if err != nil {
			return fmt.Errorf("error writing output: %w", err)
		}
		return nil
	},
}

func init() {
	generateCmd.Flags().StringToStringVar(&genFlags.vars, "vars", nil, "Provide vars to the pipeline")
	RootCmd.AddCommand(generateCmd)
}
