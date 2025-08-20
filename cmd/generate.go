/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"path/filepath"

	"github.com/chancez/yamlforge/pkg/config"
	"github.com/chancez/yamlforge/pkg/generator"
	"github.com/spf13/cobra"
)

type GenerateFlags struct {
	vars  map[string]string
	debug bool
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
		vars := make(map[string]any)
		for varName, varVal := range genFlags.vars {
			vars[varName] = varVal
		}

		cfg, err := config.ParseFile(forgeFile)
		if err != nil {
			return fmt.Errorf("error parsing pipeline %s: %w", forgeFile, err)
		}

		dir := filepath.Dir(forgeFile)
		refStore := generator.NewStore(vars)
		state := generator.NewPipeline(dir, cfg.PipelineGenerator, refStore, genFlags.debug)
		result, err := state.Generate(cmd.Context())
		if err != nil {
			return err
		}

		resultBytes, err := generator.ConvertToBytes(result)
		if err != nil {
			return err
		}

		_, err = cmd.OutOrStdout().Write(resultBytes)
		if err != nil {
			return fmt.Errorf("error writing output: %w", err)
		}
		return nil
	},
}

func init() {
	generateCmd.Flags().StringToStringVar(&genFlags.vars, "vars", nil, "Provide vars to the pipeline")
	generateCmd.Flags().BoolVar(&genFlags.debug, "debug", false, "If true, log each stage as it executes")
	RootCmd.AddCommand(generateCmd)
}
