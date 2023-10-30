// Code generated by gen.py. DO NOT EDIT.

package expconf

import (
	"github.com/santhosh-tekuri/jsonschema/v2"

	"github.com/determined-ai/determined/cluster/pkg/schemas"
)

func (h HyperparametersV0) ParsedSchema() interface{} {
	return schemas.ParsedHyperparametersV0()
}

func (h HyperparametersV0) SanityValidator() *jsonschema.Schema {
	return schemas.GetSanityValidator("http://determined.ai/schemas/expconf/v0/hyperparameters.json")
}

func (h HyperparametersV0) CompletenessValidator() *jsonschema.Schema {
	return schemas.GetCompletenessValidator("http://determined.ai/schemas/expconf/v0/hyperparameters.json")
}
