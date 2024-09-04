/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/chancez/yamlforge/pkg/config/schema"
)

var jsonSchemaCmd = &cobra.Command{
	Use:   "json-schema",
	Short: "Print the JSON schema for the forge configuration file.",
	RunE: func(cmd *cobra.Command, args []string) error {
		data, err := json.MarshalIndent(schema.Schema, "", "  ")
		if err != nil {
			return err
		}
		fmt.Println(string(data))
		return nil
	},
}

func init() {
	RootCmd.AddCommand(jsonSchemaCmd)
}
