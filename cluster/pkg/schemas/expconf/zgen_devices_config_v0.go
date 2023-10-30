// Code generated by gen.py. DO NOT EDIT.

package expconf

import (
	"github.com/santhosh-tekuri/jsonschema/v2"

	"github.com/determined-ai/determined/cluster/pkg/schemas"
)

func (d DevicesConfigV0) ParsedSchema() interface{} {
	return schemas.ParsedDevicesConfigV0()
}

func (d DevicesConfigV0) SanityValidator() *jsonschema.Schema {
	return schemas.GetSanityValidator("http://determined.ai/schemas/expconf/v0/devices.json")
}

func (d DevicesConfigV0) CompletenessValidator() *jsonschema.Schema {
	return schemas.GetCompletenessValidator("http://determined.ai/schemas/expconf/v0/devices.json")
}
