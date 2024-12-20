package json

import (
	"hit.edu/framework/pkg/apis/runtime"
	"io"
)

// TODO: Json序列化
type Serializer struct {
}

var _ runtime.Serializer = &Serializer{}

// 默认使用Json
// Yaml需要先序列化为Json
func (s *Serializer) Decode(data []byte, into runtime.Object) (runtime.Object, error) {
	return nil, nil
}

func (s *Serializer) Encode(obj runtime.Object, w io.Writer) error {
	return nil
}
