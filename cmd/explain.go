/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"bytes"
	"fmt"
	"io"
	"strings"

	"github.com/chancez/yamlforge/pkg/config/schema"
	"github.com/invopop/jsonschema"

	"github.com/spf13/cobra"
)

var explainCmd = &cobra.Command{
	Use:   "explain",
	Short: "Describe fields and structure of forge configurations.",
	Long: `This command describes the fields associated with each supported config option. Fields are identified via a simple JSONPath identifier:

        <type>.<fieldName>[.<fieldName>]
Examples:
  # Get the documentation of the config
  yfg explain config

  # Get the documentation of a nested field
  yfg explain config.pipeline.generator.helm
`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		input := strings.Split(args[0], ".")
		ty := input[0]
		fields := input[1:]

		var typeSchema *jsonschema.Schema
		var typeName string
		for defName, defValue := range schema.Schema.Definitions {
			if strings.EqualFold(defName, ty) {
				typeSchema = defValue
				typeName = defName
				break
			}
		}
		if typeSchema == nil {
			return fmt.Errorf("invalid type: %s", ty)
		}

		fieldSchema := typeSchema
		var fieldName string
		var fieldDescription string
		var fieldType string
		if len(fields) != 0 {
			// Lookup the field the user specified
			for _, field := range fields {
				foundField := false
				for pair := fieldSchema.Properties.Oldest(); pair != nil; pair = pair.Next() {
					if strings.ToLower(pair.Key) != field {
						continue
					}

					foundField = true
					fieldName = pair.Key
					fieldSchema = pair.Value
					fieldDescription = fieldSchema.Description
					fieldType = schemaTypeString(fieldSchema)

					// Now lookup the definition for this field if needed.
					// Resolve refs if they exist
					for fieldSchema.Ref != "" || fieldSchema.Items != nil {
						if fieldSchema.Ref != "" {
							var err error
							_, fieldSchema, err = getDefinition(fieldSchema.Ref)
							if err != nil {
								return err
							}
						}

						// If it's a list then lookup the inner type of the list is what we
						// need to lookup the fields.
						if fieldSchema.Items != nil {
							fieldSchema = fieldSchema.Items
						}
					}
					break
				}
				if !foundField {
					return fmt.Errorf("error: field %q does not exist", field)
				}
			}
		}

		var buf bytes.Buffer
		log := func(format string, a ...any) {
			if format != "" {
				io.WriteString(&buf, fmt.Sprintf(format, a...))
			}
			io.WriteString(&buf, "\n")
		}

		logDescription := func(s string) {
			wrapLength := 79
			io.WriteString(&buf, WrapAndIndent(s, wrapLength, 4))
			io.WriteString(&buf, "\n")
		}

		log("TYPE:\t%s", typeName)
		log("")

		if typeSchema.Description != "" {
			log("DESCRIPTION:")
			logDescription(typeSchema.Description)
			log("")
		}

		if fieldSchema != nil && fieldSchema != typeSchema {
			log("FIELD: %s <%s>", fieldName, fieldType)
			logDescription(fieldDescription)
			log("")
		}

		props := fieldSchema.Properties
		if props.Len() != 0 {
			log("FIELDS:")
			for pair := props.Oldest(); pair != nil; pair = pair.Next() {
				name := pair.Key
				schema := pair.Value
				desc := schema.Description
				ty := schemaTypeString(schema)
				required := false
				for _, req := range fieldSchema.Required {
					if req != name {
						continue
					}
					required = true
					break
				}
				if required {
					log("  %s\t<%s> -required-", name, ty)
				} else {
					log("  %s\t<%s>", name, ty)
				}
				if desc != "" {
					logDescription(desc)
				}
				log("")
			}
		}

		fmt.Println(buf.String())

		return nil
	},
}

func init() {
	RootCmd.AddCommand(explainCmd)
}

func schemaTypeString(schema *jsonschema.Schema) string {
	ty := schema.Type
	array := false
	if schema.Type == "array" {
		array = true
		if schema.Items != nil {
			schema = schema.Items
		}
		ty = schema.Type
	}
	if schema.Ref != "" {
		ty = strings.TrimPrefix(schema.Ref, "#/$defs/")
	}
	if array {
		ty = "[]" + ty
	}
	return ty
}

func getDefinition(definition string) (string, *jsonschema.Schema, error) {
	shortDefName := strings.TrimPrefix(definition, "#/$defs/")
	for defName, defValue := range schema.Schema.Definitions {
		if defName != shortDefName {
			continue
		}
		return defName, defValue, nil
	}
	return "", nil, fmt.Errorf("unable to find definition for %s", shortDefName)
}

// WrapAndIndent wraps the input text at the specified line length and indents each line with the specified number of spaces.
func WrapAndIndent(text string, lineLength int, indentSpaces int) string {
	var result strings.Builder
	words := strings.Fields(text)
	indent := strings.Repeat(" ", indentSpaces)

	currentLineLength := 0

	for _, word := range words {
		if currentLineLength+len(word)+1 > lineLength {
			result.WriteString("\n" + indent)
			currentLineLength = indentSpaces
		} else if currentLineLength > 0 {
			result.WriteString(" ")
			currentLineLength++
		}
		result.WriteString(word)
		currentLineLength += len(word)
	}

	return indent + result.String()
}
