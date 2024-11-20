package fun

type Ctx struct {
	Ip         string
	Id         string
	State      map[string]any
	RequestId  string
	MethodName string
	Send       func(id string, requestId string, data any) bool
	Close      func(id string, requestId string)
	fun        *Fun
}

type WatchClose func(callBack func())

func (Ctx Ctx) close(callBack func()) {
	conn, ok := Ctx.fun.connList.Load(Ctx.Id)
	if ok {
		on, ok := conn.(ws).onList.Load(Ctx.RequestId)
		if ok {
			on.(*onType).callBack = &callBack
		}
	}
}

type onType struct {
	callBack   *func()
	methodName string
}
