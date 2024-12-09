package runtime

import (
	"hit.edu/framework/pkg/apimachinery/runtime/schema"
	"net/url"
)

type parameterCodec struct {
}

var _ ParameterCodec = &parameterCodec{}

// DecodeParameters converts the provided url.Values into an object of type From with the kind of into, and then
// converts that object to into (if necessary). Returns an error if the operation cannot be completed.
func (c *parameterCodec) DecodeParameters(parameters url.Values, from schema.GroupVersion, into Object) error {
	panic("To impl")
}

// EncodeParameters converts the provided object into the to version, then converts that object to url.Values.
// Returns an error if conversion is not possible.
func (c *parameterCodec) EncodeParameters(obj Object, to schema.GroupVersion) (url.Values, error) {
	// TODO: 实现该逻辑
	panic("To impl")
}
