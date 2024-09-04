// This program generates schema.json
package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/invopop/jsonschema"

	"github.com/chancez/yamlforge/pkg/config"
)

func getJSONSchema() (*jsonschema.Schema, error) {
	reflector := &jsonschema.Reflector{
		CommentMap: make(map[string]string),
	}
	if err := ExtractGoComments("github.com/chancez/yamlforge", "./", reflector.CommentMap); err != nil {
		return nil, fmt.Errorf("error extracting go comments: %w", err)
	}
	return reflector.Reflect(&config.Config{}), nil
}

func writeSchemaJSON(dest string) error {
	schema, err := getJSONSchema()
	if err != nil {
		return fmt.Errorf("error getting schema: %w", err)
	}
	data, err := json.MarshalIndent(schema, "", "  ")
	if err != nil {
		return fmt.Errorf("error marshaling schema: %w", err)
	}
	err = os.WriteFile(dest, data, 0644)
	if err != nil {
		return fmt.Errorf("error writing schema.json: %w", err)
	}
	return nil
}

func main() {
	args := os.Args[1:]
	if len(args) != 1 {
		fmt.Println("must provide destination as an argument")
		os.Exit(1)
	}
	dest := args[0]
	err := writeSchemaJSON(dest)
	if err != nil {
		fmt.Printf("error when trying to generate schema: %s\n", err)
		os.Exit(1)
	}
}
