package schema

// From K8s
type ObjectKind interface {
	SetGroupVersionKind(kind GroupVersionKind)
	GroupVersionKind() GroupVersionKind
}

// From K8s
type GroupVersionKind struct {
	Group   string
	Version string
	Kind    string
}
