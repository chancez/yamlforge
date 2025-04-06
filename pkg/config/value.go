package config

import (
	"encoding/json"
	"fmt"

	"github.com/invopop/jsonschema"
)

type StringValue struct {
	String *string
	Value  *Value
}

var _ json.Unmarshaler = (*StringValue)(nil)

func (sv *StringValue) UnmarshalJSON(data []byte) error {
	// Try to unmarshal as a string.
	var s string
	if err := json.Unmarshal(data, &s); err == nil {
		sv.String = &s
		return nil
	}

	// Try to unmarshal as a Value.
	var v Value
	if err := json.Unmarshal(data, &v); err == nil {
		sv.Value = &v
		return nil
	}

	return fmt.Errorf("StringValue: cannot unmarshal %s", data)
}

func (StringValue) JSONSchema() *jsonschema.Schema {
	return oneOfTypeOrValueSchema("string")
}

type BoolValue struct {
	Bool  *bool
	Value *Value
}

var _ json.Unmarshaler = (*BoolValue)(nil)

func (bv *BoolValue) UnmarshalJSON(data []byte) error {
	// Try to unmarshal as a bool.
	var b bool
	if err := json.Unmarshal(data, &b); err == nil {
		bv.Bool = &b
		return nil
	}

	// Try to unmarshal as a Value.
	var v Value
	if err := json.Unmarshal(data, &v); err == nil {
		bv.Value = &v
		return nil
	}

	return fmt.Errorf("BoolValue: cannot unmarshal %s", data)
}

func (BoolValue) JSONSchema() *jsonschema.Schema {
	return oneOfTypeOrValueSchema("boolean")
}

func oneOfTypeOrValueSchema(typ string) *jsonschema.Schema {
	return &jsonschema.Schema{
		OneOf: []*jsonschema.Schema{
			{
				Type: typ,
			},
			{
				Ref: "#/$defs/Value",
			},
		},
	}
}

// Value provides inputs to generators.
type Value struct {
	// Var allows defining variables that can be externally provided to a pipeline.
	Var string `yaml:"var,omitempty" json:"var,omitempty" jsonschema:"oneof_required=var"`
	// Ref takes the name of a previous stage in the pipeline and returns the output of that stage.
	Ref string `yaml:"ref,omitempty" json:"ref,omitempty" jsonschema:"oneof_required=ref"`
	// File takes a path relative to this pipeline file to read and returns the content of the file specified.
	File string `yaml:"file,omitempty" json:"file,omitempty" jsonschema:"oneof_required=file"`
	// Value simply returns the value specified. It can be any valid YAML/JSON type ( string, boolean, number, array, object).
	Value any `yaml:"value,omitempty" json:"value,omitempty" jsonschema:"oneof_required=value"`
	// IgnoreMissing specifies if the generator should ignore missing references or files. If set to true, the generator will return an empty string instead of an error.
	IgnoreMissing bool `yaml:"ignoreMissing,omitempty" json:"ignoreMissing,omitempty"`
	// Default specifies the default value to use if a ref, variable, or file is
	// missing. Has no effect unless ignoreMissing is true.
	// It can be any valid YAML/JSON type ( string, boolean, number, array, object).
	Default any `yaml:"default,omitempty" json:"default,omitempty"`
}

// NamedValue is a Value with a name.
type NamedValue struct {
	// Name is the name of this variable.
	Name  string `yaml:"name" json:"name"`
	Value `yaml:",inline" json:",inline"`
}

// ParsedValue provides parsed values to generators.
type ParsedValue struct {
	// Format defines the format to parse the retrieved value as. Valid options are yaml or json.
	Format string `yaml:"format" json:"format" jsonschema:"enum=yaml,enum=json"`
	Value  `yaml:",inline" json:",inline"`
}
