package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func strPtr(s string) *string {
	return &s
}

func TestDecodeStringValue(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    StringValue
		wantErr bool
	}{
		{
			name:  "string input",
			input: "foo",
			want: StringValue{
				String: strPtr("foo"),
			},
		},
		{
			name:  "ref input",
			input: "ref: some-stage",
			want: StringValue{
				Value: &Value{
					Ref: "some-stage",
				},
			},
		},
		{
			name:    "invalid input",
			input:   "true",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got StringValue
			err := DecodeYAML([]byte(tt.input), &got)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func boolPtr(b bool) *bool {
	return &b
}

func TestDecodeBoolValue(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    BoolValue
		wantErr bool
	}{
		{
			name:  "string input",
			input: "true",
			want: BoolValue{
				Bool: boolPtr(true),
			},
		},
		{
			name:  "ref input",
			input: "ref: some-stage",
			want: BoolValue{
				Value: &Value{
					Ref: "some-stage",
				},
			},
		},
		{
			name:    "invalid input",
			input:   `"a string"`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got BoolValue
			err := DecodeYAML([]byte(tt.input), &got)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}
