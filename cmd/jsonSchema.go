/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/invopop/jsonschema"
	"github.com/spf13/cobra"

	"github.com/chancez/yamlforge/pkg/config"
)

// jsonSchemaCmd represents the jsonSchema command
var jsonSchemaCmd = &cobra.Command{
	Use:   "json-schema",
	Short: "Print the JSON schema for the forge configuration file.",
	RunE: func(cmd *cobra.Command, args []string) error {
		schema := jsonschema.Reflect(&config.Config{})
		data, err := json.MarshalIndent(schema, "", "  ")
		if err != nil {
			return err
		}
		fmt.Println(string(data))
		return nil
	},
}

func init() {
	rootCmd.AddCommand(jsonSchemaCmd)
}
