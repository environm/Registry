package watch

import "hit.edu/framework/pkg/apimachinery/runtime"

type Interface interface {
	//
	Stop()

	//
	ResultChan() <-chan Event
}

type EventType string

const (
	Added    EventType = "ADDED"
	Modified EventType = "MODIFIED"
	Deleted  EventType = "DELETED"
	Error    EventType = "ERROR"
)

var (
	DefaultChanSize int32 = 100
)

type Event struct {
	Type EventType

	// TODO: 定义Object的对象用例
	Object runtime.Object
}
