// Code generated by gen.py. DO NOT EDIT.

package expconf

import (
	"github.com/santhosh-tekuri/jsonschema/v2"

	"github.com/determined-ai/determined/master/pkg/schemas"
)

func (p ProxyPortsConfigV0) ParsedSchema() interface{} {
	return schemas.ParsedProxyPortsConfigV0()
}

func (p ProxyPortsConfigV0) SanityValidator() *jsonschema.Schema {
	return schemas.GetSanityValidator("http://determined.ai/schemas/expconf/v0/proxy-ports.json")
}

func (p ProxyPortsConfigV0) CompletenessValidator() *jsonschema.Schema {
	return schemas.GetCompletenessValidator("http://determined.ai/schemas/expconf/v0/proxy-ports.json")
}
