package fun

import "context"

type Ctx struct {
	Ip          string
	Id          string
	State       map[string]string
	RequestId   string
	MethodName  string
	ServiceName string
	Send        func(id string, requestId string, data any) bool
	Close       func(id string, requestId string)
	fun         *Fun
}

type Proxy struct {
	Open   *func()
	Close  *func()
	Ctx    context.Context
	Cancel context.CancelFunc
}

type onType struct {
	serviceName string
	methodName  string
	proxy       *Proxy
}
