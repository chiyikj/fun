package fun

type Ctx struct {
	Ip          string
	Id          string
	State       map[string]string
	RequestId   string
	MethodName  string
	ServiceName string
	Push        func(id string, requestId string, data any) bool
	Close       func(id string, requestId string)
	fun         *Fun
}

type ProxyClose func(callBack func())

type onType struct {
	serviceName string
	methodName  string
	callBack    *func()
}
