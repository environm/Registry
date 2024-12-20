package runtime

import "hit.edu/framework/pkg/apimachinery/runtime/schema"

func (obj *TypeMeta) GetObjectKind() schema.ObjectKind { return obj }
