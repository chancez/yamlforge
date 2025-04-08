package config

import (
	"bytes"
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
	// Unmarshal into Value
	ok, err := maybeUnmarshalValue(data, &sv.Value)
	if err != nil {
		return err
	}
	if ok {
		return nil
	}

	// Try to unmarshal as a string.
	var s string
	if err := json.Unmarshal(data, &s); err == nil {
		sv.String = &s
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
	// Unmarshal into Value
	ok, err := maybeUnmarshalValue(data, &bv.Value)
	if err != nil {
		return err
	}
	if ok {
		return nil
	}

	// Try to unmarshal as a bool.
	var b bool
	if err := json.Unmarshal(data, &b); err == nil {
		bv.Bool = &b
		return nil
	}

	return fmt.Errorf("BoolValue: cannot unmarshal %s", data)
}

func (BoolValue) JSONSchema() *jsonschema.Schema {
	return oneOfTypeOrValueSchema("boolean")
}

func maybeUnmarshalValue(data []byte, val **Value) (bool, error) {
	trimmed := bytes.TrimSpace(data)
	if len(trimmed) == 0 {
		return false, nil
	}

	unmarshalValue := func() error {
		*val = new(Value)
		// Unmarshal into Value.
		if err := json.Unmarshal(data, *val); err != nil {
			return err
		}
		return nil
	}

	// If the JSON starts with '{', check if it contains any keys that belong to Value.
	if trimmed[0] == '{' {
		var obj map[string]json.RawMessage
		if err := json.Unmarshal(data, &obj); err != nil {
			return false, err
		}
		if _, hasVar := obj["var"]; hasVar {
			return true, unmarshalValue()
		}
		if _, hasRef := obj["ref"]; hasRef {
			return true, unmarshalValue()
		}
		if _, hasFile := obj["file"]; hasFile {
			return true, unmarshalValue()
		}
		if _, hasValue := obj["value"]; hasValue {
			return true, unmarshalValue()
		}
		// If none of the specific keys exist, then it's not a "Value" type, so return false without decoding anything
	}
	return false, nil
}

type MapValue struct {
	Map   map[string]any
	Value *Value
}

var _ json.Unmarshaler = (*MapValue)(nil)

func (mv *MapValue) UnmarshalJSON(data []byte) error {
	// Unmarshal into Value
	ok, err := maybeUnmarshalValue(data, &mv.Value)
	if err != nil {
		return err
	}
	if ok {
		return nil
	}

	// Try to unmarshal as a generic map.
	var m map[string]any
	if err := json.Unmarshal(data, &m); err == nil {
		mv.Map = m
		return nil
	}

	return fmt.Errorf("MapValue: cannot unmarshal %s", data)
}

func (MapValue) JSONSchema() *jsonschema.Schema {
	return oneOfTypeOrValueSchema("object")
}

type AnyValue struct {
	Any   *any
	Value *Value
}

var _ json.Unmarshaler = (*AnyValue)(nil)

func (av *AnyValue) UnmarshalJSON(data []byte) error {
	// Unmarshal into Value
	ok, err := maybeUnmarshalValue(data, &av.Value)
	if err != nil {
		return err
	}
	if ok {
		return nil
	}
	// Unmarshal into any.
	var anyVal any
	if err := json.Unmarshal(data, &anyVal); err != nil {
		return err
	}
	av.Any = &anyVal
	return nil
}

func (AnyValue) JSONSchema() *jsonschema.Schema {
	return oneOfTypeOrValueSchema("number", "string", "boolean", "null", "object", "array")
}

func oneOfTypeOrValueSchema(typs ...string) *jsonschema.Schema {
	var schemas []*jsonschema.Schema
	for _, typ := range typs {
		schemas = append(schemas, &jsonschema.Schema{
			Type: typ,
		})
	}
	schemas = append(schemas,
		&jsonschema.Schema{Ref: "#/$defs/Value"},
	)
	return &jsonschema.Schema{
		OneOf: schemas,
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
	// Format defines the format to parse the retrieved value as. Valid options
	// are yaml or json. Defaults to yaml if unspecified, or if the value
	// references a file, it will attempt to use the file extension to determine
	// the correct format.
	Format string `yaml:"format" json:"format" jsonschema:"enum=yaml,enum=json,default=yaml"`
	Value  `yaml:",inline" json:",inline"`
}

func (pv ParsedValue) String() string {
	b, _ := json.Marshal(pv)
	return string(b)
}
