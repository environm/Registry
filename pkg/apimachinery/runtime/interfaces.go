package runtime

import (
	"hit.edu/framework/pkg/apimachinery/runtime/schema"
	"io"
	"net/url"
)

type Object interface {
	GetObjectKind() schema.ObjectKind
	DeepCopyObject() Object
}

type Serializer interface {
	Encoder
	Decoder
}

type Encoder interface {
	Encode(obj Object, writer io.Writer) error
}

type Decoder interface {
	Decode(data []byte, into Object) (Object, error)
}

type Codec Serializer

type ParameterCodec interface {
	// DecodeParameters takes the given url.Values in the specified group version and decodes them
	// into the provided object, or returns an error.
	DecodeParameters(parameters url.Values, from schema.GroupVersion, into Object) error
	// EncodeParameters encodes the provided object as query parameters or returns an error.
	EncodeParameters(obj Object, to schema.GroupVersion) (url.Values, error)
}
