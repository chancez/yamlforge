/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"bytes"
	"fmt"
	"io"
	"slices"
	"strings"

	"github.com/chancez/yamlforge/pkg/config/schema"
	"github.com/invopop/jsonschema"

	"github.com/spf13/cobra"
)

type ExplainFlags struct {
	verbose bool
}

var explainFlags ExplainFlags

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
		input := strings.SplitN(args[0], ".", 2)
		ty := input[0]
		var field string
		if len(input) == 2 {
			field = input[1]
			if field == "" {
				return fmt.Errorf("error: field name is empty")
			}
		}

		typeSchema, typeName, err := getTypeSchema(ty)
		if err != nil {
			return err
		}

		fieldSchema := typeSchema
		var fieldName string
		if field != "" {
			// Lookup the field the user specified
			fieldSchema, fieldName, err = getFieldSchema(field, typeSchema)
			if err != nil {
				return err
			}

		}

		var fieldDescription string
		var fieldType string
		if fieldSchema != nil {
			// Store the description and type of the field
			fieldDescription = fieldSchema.Description
			fieldType = schemaTypeString(fieldSchema)
		}

		var buf bytes.Buffer

		bufLog(&buf, "TYPE:\t%s", typeName)
		bufLog(&buf, "")

		if typeSchema.Description != "" {
			bufLog(&buf, "DESCRIPTION:")
			bufLogIndentWrapped(&buf, typeSchema.Description)
			bufLog(&buf, "")
		}

		fieldSchema, _, err = getSubSchema(fieldSchema)
		if err != nil {
			return err
		}
		if fieldSchema != nil && fieldSchema != typeSchema {
			bufLog(&buf, "FIELD: %s <%s>", fieldName, fieldType)
			bufLogIndentWrapped(&buf, fieldDescription)
			bufLog(&buf, "")

			// Terminal field, no sub-fields, and the field is a non-basic type.
			// Lookup the type of the field and print the description if it exists.
			if !isBasicType(fieldType) && fieldSchema.Properties.Len() == 0 {
				bufLog(&buf, "FIELD TYPE:\t%s <%s>", fieldType, schemaTypeString(fieldSchema))
				if fieldSchema.Description != "" {
					bufLogIndentWrapped(&buf, fieldSchema.Description)
					bufLog(&buf, "")
				}
				if explainFlags.verbose && len(fieldSchema.OneOf) != 0 {
					err := logSubTypes(&buf, fieldSchema.OneOf)
					if err != nil {
						return err
					}
				} else {
					bufLogIndentWrapped(&buf, "For details run 'yfg explain %s' or re-run the previous command with --verbose.", fieldType)
				}
			}
		}

		if fieldSchema.Properties.Len() != 0 {
			bufLog(&buf, "FIELDS:")
			logSchemaProperties(&buf, fieldSchema)
		}

		fmt.Println(buf.String())

		return nil
	},
}

func init() {
	explainCmd.Flags().BoolVar(&explainFlags.verbose, "verbose", false, "Enable verbose output")
	RootCmd.AddCommand(explainCmd)
}

func schemaTypeString(schema *jsonschema.Schema) string {
	if schema.OneOf != nil {
		var tys []string
		for _, s := range schema.OneOf {
			tys = append(tys, schemaTypeString(s))
		}
		return "oneOf(" + strings.Join(tys, ", ") + ")"
	}
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
	if array && ty != "array" {
		ty = "[]" + ty
	}
	return ty
}

func isBasicType(typ string) bool {
	return slices.Contains([]string{
		"string", "number", "integer", "boolean", "null", "object", "array",
	}, typ)
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

func getTypeSchema(typeName string) (*jsonschema.Schema, string, error) {
	for defName, defValue := range schema.Schema.Definitions {
		if strings.EqualFold(defName, typeName) {
			return defValue, defName, nil
		}
	}
	return nil, "", fmt.Errorf("unable to find definition for %s", typeName)
}

func getFieldSchema(field string, typeSchema *jsonschema.Schema) (*jsonschema.Schema, string, error) {
	fieldSchema := typeSchema
	subSchema := fieldSchema
	fields := strings.Split(field, ".")
	var fieldName string

	for _, field := range fields {
		foundField := false
		for pair := subSchema.Properties.Oldest(); pair != nil; pair = pair.Next() {
			if !strings.EqualFold(pair.Key, field) {
				continue
			}
			foundField = true
			fieldName = pair.Key
			fieldSchema = pair.Value
			// Still having troublehere because it returns the sub type and we need both the parent and the sub-type.
			var err error
			subSchema, _, err = getSubSchema(fieldSchema)
			if err != nil {
				return nil, "", err
			}
			break
		}
		if !foundField {
			return nil, "", fmt.Errorf("error: field %q does not exist", field)
		}
	}
	return fieldSchema, fieldName, nil
}

func getSubSchema(schema *jsonschema.Schema) (*jsonschema.Schema, string, error) {
	// Now lookup the definition for this field if needed.
	// Resolve refs if they exist
	var typeName string
	for schema.Ref != "" || schema.Items != nil {
		if schema.Ref != "" {
			var err error
			typeName = strings.TrimPrefix(schema.Ref, "#/$defs/")
			schema, _, err = getTypeSchema(typeName)
			if err != nil {
				return nil, "", err
			}
		}

		// If it's a list then lookup the inner type of the list is what we
		// need to lookup the fields.
		if schema.Items != nil {
			schema = schema.Items
		}
	}
	return schema, typeName, nil
}

func bufLog(buf *bytes.Buffer, format string, a ...any) {
	if format != "" {
		//nolint:errcheck
		io.WriteString(buf, fmt.Sprintf(format, a...))
	}
	//nolint:errcheck
	io.WriteString(buf, "\n")
}

func bufLogIndentWrapped(buf *bytes.Buffer, format string, a ...any) {
	s := fmt.Sprintf(format, a...)
	wrapLength := 79
	//nolint:errcheck
	io.WriteString(buf, WrapAndIndent(s, wrapLength, 4))
	//nolint:errcheck
	io.WriteString(buf, "\n")
}

func logSchemaProperties(buf *bytes.Buffer, schema *jsonschema.Schema) {
	for pair := schema.Properties.Oldest(); pair != nil; pair = pair.Next() {
		name := pair.Key
		schema := pair.Value
		desc := schema.Description
		ty := schemaTypeString(schema)
		required := false
		for _, req := range schema.Required {
			if req != name {
				continue
			}
			required = true
			break
		}
		if required {
			bufLog(buf, "  %s\t<%s> -required-", name, ty)
		} else {
			bufLog(buf, "  %s\t<%s>", name, ty)
		}
		if desc != "" {
			bufLogIndentWrapped(buf, desc)
		}
		bufLog(buf, "")
	}
}

func logSubTypes(buf *bytes.Buffer, schemas []*jsonschema.Schema) error {
	for _, schema := range schemas {
		if schema.Required != nil || isBasicType(schema.Type) {
			continue
		}
		var subType string
		subSchema, subType, err := getSubSchema(schema)
		if err != nil {
			return err
		}
		bufLog(buf, "")
		bufLog(buf, "SUB TYPE:\t%s", subType)
		if subSchema.Description != "" {
			bufLogIndentWrapped(buf, subSchema.Description)
		}
		bufLog(buf, "")
		if subSchema.Properties.Len() != 0 {
			bufLog(buf, "SUB TYPE FIELDS:")
			bufLog(buf, "")
			logSchemaProperties(buf, subSchema)
		}
	}
	return nil
}
