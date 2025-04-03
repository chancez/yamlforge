package config

import (
	"bytes"
	"io"

	"github.com/goccy/go-yaml"
)

var DefaultYAMLDecodeOptions = []yaml.DecodeOption{
	yaml.Strict(),
	yaml.UseJSONUnmarshaler(),
}

var DefaultYAMLEncodeOptions = NewYAMLEncodeOptions(0)

func NewYAMLEncodeOptions(indent int) []yaml.EncodeOption {
	if indent == 0 {
		indent = 4
	}
	return []yaml.EncodeOption{
		yaml.Indent(indent),
		yaml.IndentSequence(true),
		yaml.UseJSONMarshaler(),
	}
}

func NewYAMLDecoder(r io.Reader) *yaml.Decoder {
	return yaml.NewDecoder(r, DefaultYAMLDecodeOptions...)
}

func NewYAMLEncoder(w io.Writer) *yaml.Encoder {
	return yaml.NewEncoder(w, DefaultYAMLEncodeOptions...)
}

func NewYAMLEncoderWithIndent(w io.Writer, indent int) *yaml.Encoder {
	return yaml.NewEncoder(w, NewYAMLEncodeOptions(indent)...)
}

func DecodeYAML(data []byte, v any) error {
	dec := NewYAMLDecoder(bytes.NewBuffer(data))
	return dec.Decode(v)
}

func EncodeYAML(v any) ([]byte, error) {
	return yaml.MarshalWithOptions(v, DefaultYAMLEncodeOptions...)
}
