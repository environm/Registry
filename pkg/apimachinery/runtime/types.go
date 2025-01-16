package runtime

import "hit.edu/framework/pkg/apimachinery/runtime/schema"

type TypeMeta struct {
	//
	Kind string
	//
	APIVersion string
}

func (obj *TypeMeta) SetGroupVersionKind(kind schema.GroupVersionKind) {
	//TODO implement me
	panic("implement me")
}

func (obj *TypeMeta) GroupVersionKind() schema.GroupVersionKind {
	//TODO implement me
	panic("implement me")
}
