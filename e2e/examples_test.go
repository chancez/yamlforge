package e2e

import (
	"bytes"
	"path"
	"strings"
	"testing"

	"github.com/chancez/yamlforge/cmd"
	"github.com/stretchr/testify/require"
)

func TestExamples(t *testing.T) {
	trim := func(s string) string {
		return strings.TrimLeft(s, "\n")
	}
	tests := []struct {
		file     string
		expected string
	}{
		{
			file: "file.yfg.yaml",
			expected: trim(`
foo:
  bar: baz
  key: |
    some value
  another_key: |
    {{ .SomeValue }}
`),
		},
		{
			file: "exec.yfg.yaml",
			expected: trim(`
foo:
    bar: asdf hjkl
`),
		},
		{
			file: "jq.yfg.yaml",
			expected: trim(`
another_key: |
    {{ .SomeValue }}
bar: baz
key: |
    some value
`),
		},
		{
			file: "merge.yfg.yaml",
			expected: trim(`
foo:
    another_key: |
        {{ .SomeValue }}
    bar: asdf hjkl
    key: |
        some value
`),
		},
		{
			file: "template.yfg.yaml",
			expected: trim(`
foo:
  bar: baz
  key: |
    some value
  another_key: |
    dog
`),
		},
		{
			file: "template-literal.yfg.yaml",
			expected: trim(`
some-key: some-value
`),
		},
		{
			file: "single-generator.yfg.yaml",
			expected: trim(`
foo: bar
key: value
`),
		},
		{
			file: "reusable-transformer.yfg.yaml",
			expected: trim(`
foo:
    another_key: |
        {{ .SomeValue }}
    bar: baz
    key: |
        some value
some-new-key: hello world
`),
		},
	}
	for _, tt := range tests {
		t.Run(tt.file, func(t *testing.T) {
			p := path.Join("../examples", tt.file)
			require.FileExists(t, p, "example file must exist")
			var buf bytes.Buffer
			c := cmd.RootCmd
			c.SetArgs([]string{"generate", p})
			c.SetOut(&buf)
			err := c.Execute()
			require.NoError(t, err, "yfg generate should succeed on examples")
			require.Equal(t, tt.expected, buf.String(), "example output should match expected")
		})
	}
}
