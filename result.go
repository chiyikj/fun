package fun

import "reflect"

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
	var resultData *any
	if data != nil {
		// 检查是否为切片类型
		dataType := reflect.TypeOf(data)
		if dataType.Kind() == reflect.Slice {
			// 检查是否为空切片
			dataValue := reflect.ValueOf(data)
			if dataValue.Len() == 0 {
				// 创建一个新的空切片而不是nil
				emptySlice := reflect.MakeSlice(dataValue.Type(), 0, 0).Interface()
				resultData = &emptySlice
			} else {
				resultData = &data
			}
		} else {
			resultData = &data
		}
	}
	return Result[any]{Data: resultData, Status: successCode}
}

func Error(code uint16, msg string) Result[any] {
	return Result[any]{Code: &code, Msg: &msg, Status: errorCode}
}

func callError(msg string) Result[any] {
	return Result[any]{Msg: &msg, Status: cellErrorCode}
}

func closeError(requestId string) Result[any] {
	return Result[any]{Id: requestId, Status: closeErrorCode}
}
