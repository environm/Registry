package apis

import "hit.edu/framework/pkg/apimachinery/runtime/schema"

const GroupName = ""

// Resource takes an unqualified resource and returns a Group qualified GroupResource
func Resource(resource string) schema.GroupResource {
	return schema.GroupResource{Group: GroupName, Resource: resource}
}
