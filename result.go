package fun

const (
	successCode = iota
	cellErrorCode
	errorCode
	closeErrorCode
)

type Result[T any] struct {
	Id     string
	Code   *uint16
	Data   *T
	Msg    *string
	Status uint8
}

func success(data any) Result[any] {
	return Result[any]{Data: &data, Status: successCode}
}

func Error(code uint16, msg string) Result[any] {
	return Result[any]{Code: &code, Msg: &msg, Status: errorCode}
}

func callError(msg string) Result[any] {
	return Result[any]{Msg: &msg, Status: cellErrorCode}
}

func closeError(requestId string) Result[any] {
	return Result[any]{Id: requestId, Status: CloseErrorCode}
}
