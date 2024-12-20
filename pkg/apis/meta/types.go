package meta

import (
	"fmt"
	"hit.edu/framework/pkg/apimachinery/runtime"
)

type ObjectMeta struct {
	//
	Name string
	//
	Namespace string
}

// 输出单个资源
type APIResource struct {
	Name string

	Group string

	Version string

	Kind string

	Verbs Verbs

	ShortName []string

	// TODO: 数据一致性相关字段
}

type Verbs []string

func (vs Verbs) String() string {
	return fmt.Sprintf("%v", []string(vs))
}

type APIResourceList struct {
	runtime.TypeMeta

	//
	GroupVersion string
	//
	APIResources []APIResource
}

// TODO: Label Selector
// TODO: 输出格式定义
