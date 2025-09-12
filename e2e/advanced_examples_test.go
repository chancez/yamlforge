//go:build advanced_examples

package e2e

import (
	"bytes"
	"embed"
	"path"
	"testing"

	"github.com/chancez/yamlforge/cmd"
	"github.com/stretchr/testify/require"
)

var (
	//go:embed testdata/*
	testdata embed.FS
)

func TestAdvancedExamples(t *testing.T) {
	readFile := func(name string) []byte {
		t.Helper()
		d, err := testdata.ReadFile(name)
		require.NoError(t, err)
		return d
	}
	tests := []struct {
		file     string
		expected string
	}{
		{
			file:     "helm.yfg.yaml",
			expected: string(readFile("testdata/helm.txt")),
		},
		{
			file:     "advanced/helm-templated-values.yfg.yaml",
			expected: string(readFile("testdata/helm-templated-values.txt")),
		},
		{
			file:     "advanced/dynamic-pipeline.yfg.yaml",
			expected: string(readFile("testdata/dynamic-pipeline.txt")),
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
