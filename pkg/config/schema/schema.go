package schema

import (
	_ "embed"
	"encoding/json"
	"fmt"

	"github.com/invopop/jsonschema"
)

//go:embed schema.json
var schemaJSON []byte

var Schema *jsonschema.Schema

func init() {
	Schema = new(jsonschema.Schema)
	err := json.Unmarshal(schemaJSON, Schema)
	if err != nil {
		panic(fmt.Sprintf("error parsing schema.json: %s", err))
	}
}
