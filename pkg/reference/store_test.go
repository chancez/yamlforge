package reference

import (
	"os"
	"path"
	"testing"

	"github.com/chancez/yamlforge/pkg/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStore(t *testing.T) {
	store := NewStore(map[string][]byte{
		"some-var": []byte(`var-data`),
	})

	err := store.AddReference("example", []byte(`ref-data`))
	require.NoError(t, err)

	// Duplicates are not allowed
	err = store.AddReference("example", []byte(``))
	require.Error(t, err)

	// Looking up references
	refData, err := store.GetValueBytes("", config.Value{
		Ref: "example",
	})
	require.NoError(t, err)
	assert.Equal(t, []byte(`ref-data`), refData)

	// Invalid refs should return an error
	_, err = store.GetValueBytes("", config.Value{
		Ref: "does not exist",
	})
	require.Error(t, err)

	// Test variables lookup
	varData, err := store.GetValueBytes("", config.Value{
		Var: "some-var",
	})
	require.NoError(t, err)
	assert.Equal(t, []byte(`var-data`), varData)

	// Invalid variables should return an error
	_, err = store.GetValueBytes("", config.Value{
		Var: "does not exist",
	})
	require.Error(t, err)

	// Values should be returned as their YAML encoded value
	valData, err := store.GetValueBytes("", config.Value{
		Value: "string-val",
	})
	require.NoError(t, err)
	assert.Equal(t, []byte(`string-val`), valData)

	valData2, err := store.GetValueBytes("", config.Value{
		Value: true,
	})
	require.NoError(t, err)
	trueBytes, err := config.EncodeYAML(true)
	require.NoError(t, err)
	assert.Equal(t, trueBytes, valData2)

	tmpDir := t.TempDir()
	err = os.WriteFile(path.Join(tmpDir, "example.txt"), []byte(`some-file-data`), 0640)
	require.NoError(t, err)

	fileData, err := store.GetValueBytes(tmpDir, config.Value{
		File: "example.txt",
	})
	require.NoError(t, err)
	assert.Equal(t, []byte(`some-file-data`), fileData)

	// Look up an existing ref as a string
	strData, err := store.GetStringValue("", config.StringValue{
		Value: &config.Value{
			Ref: "example",
		},
	})
	require.NoError(t, err)
	assert.Equal(t, "ref-data", strData)

	// Call GetStringValue on a non-ref value
	exampleStr := "example-str"
	strData2, err := store.GetStringValue("", config.StringValue{
		String: &exampleStr,
	})
	require.NoError(t, err)
	assert.Equal(t, exampleStr, strData2)

	// Store a boolean as a ref
	err = store.AddReference("bool-ref", []byte(`true`))
	require.NoError(t, err)

	// Look up an existing ref as a bool
	boolData, err := store.GetBoolValue("", config.BoolValue{
		Value: &config.Value{
			Ref: "bool-ref",
		},
	})
	require.NoError(t, err)
	assert.Equal(t, true, boolData)

	// Call GetBoolValue on a non-ref value
	exampleBool := true
	boolData2, err := store.GetBoolValue("", config.BoolValue{
		Bool: &exampleBool,
	})
	require.NoError(t, err)
	assert.Equal(t, true, boolData2)
}
