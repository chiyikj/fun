package fun

const (
	success = iota
	cellError
	returnError
)

type Result struct {
	Id   string
	Code *uint16
	//通用的类型修饰 在error的情况下表示异常
	Data   any
	status uint8
}

func resultSuccess(data any) Result {
	return Result{Data: data, status: success}
}

func Error(code uint16, msg string) Result {
	return Result{Code: &code, Data: msg, status: returnError}
}

func resultCallError(msg any) Result {
	return Result{Data: msg, status: cellError}
}
