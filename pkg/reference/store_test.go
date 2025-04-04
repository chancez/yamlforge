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
	refData, err := store.GetReference("", config.Value{
		Ref: "example",
	})
	require.NoError(t, err)
	assert.Equal(t, []byte(`ref-data`), refData)

	// Invalid refs should return an error
	_, err = store.GetReference("", config.Value{
		Ref: "does not exist",
	})
	require.Error(t, err)

	// Test variables lookup
	varData, err := store.GetReference("", config.Value{
		Var: "some-var",
	})
	require.NoError(t, err)
	assert.Equal(t, []byte(`var-data`), varData)

	// Invalid variables should return an error
	_, err = store.GetReference("", config.Value{
		Var: "does not exist",
	})
	require.Error(t, err)

	// Values should be returned as their YAML encoded value
	valData, err := store.GetReference("", config.Value{
		Value: "string-val",
	})
	require.NoError(t, err)
	assert.Equal(t, []byte(`string-val`), valData)

	valData2, err := store.GetReference("", config.Value{
		Value: true,
	})
	require.NoError(t, err)
	trueBytes, err := config.EncodeYAML(true)
	require.NoError(t, err)
	assert.Equal(t, trueBytes, valData2)

	tmpDir := t.TempDir()
	err = os.WriteFile(path.Join(tmpDir, "example.txt"), []byte(`some-file-data`), 0640)
	require.NoError(t, err)

	fileData, err := store.GetReference(tmpDir, config.Value{
		File: "example.txt",
	})
	require.NoError(t, err)
	assert.Equal(t, []byte(`some-file-data`), fileData)
}
