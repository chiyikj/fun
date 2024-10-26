package fun

const (
	Success = iota
	CellError
	ReturnError
	CloseError
)

type Result struct {
	Id   string
	Code uint16
	//通用的类型修饰 在error的情况下表示异常
	Data   any
	Status uint8
}

func resultSuccess(data any) Result {
	return Result{Data: data, Status: Success}
}

func Error(code uint16, msg string) Result {
	return Result{Code: code, Data: msg, Status: ReturnError}
}

func resultCallError(msg any) Result {
	return Result{Data: msg, Status: CellError}
}
