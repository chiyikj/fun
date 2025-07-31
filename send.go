package fun

import (
	"github.com/gorilla/websocket"
	"reflect"
)

// 普通发送信息
func (fun *Fun) send(id string, text any) bool {
	return fun.getConnInfoAndSend(id, func(loadConnInfo connInfoType) error {
		return loadConnInfo.conn.WriteJSON(text)
	})
}

// 推送
func (fun *Fun) Push(id string, requestId string, data any) bool {
	connInfo, ok := fun.connList.Load(id)
	if !ok {
		return false
	}
	loadConnInfo := connInfo.(connInfoType)
	loadConnInfo.mu.Lock()
	defer loadConnInfo.mu.Unlock()
	var result Result[any]
	result.Id = requestId
	on, ok := loadConnInfo.onList.Load(requestId)
	if ok {
		method := fun.serviceList[on.(onType).serviceName].methodList[on.(onType).methodName]
		if method.method.Type.Out(0).Elem() == reflect.TypeOf(data) {
			result = success(data)
		} else {
			result = callError("fun:" + on.(onType).methodName + " method return type Inconsistent")
		}
		loadConnInfo.onList.Delete(requestId)
	} else {
		return false
	}
	err := loadConnInfo.conn.WriteJSON(result)
	if err != nil {
		return false
	}
	return true
}

// 发送二进制信息
func (fun *Fun) sendBinary(id string, data []byte) bool {
	return fun.getConnInfoAndSend(id, func(loadConnInfo connInfoType) error {
		return loadConnInfo.conn.WriteMessage(websocket.BinaryMessage, data)
	})
}

func (fun *Fun) sendPong(id string) {
	fun.sendBinary(id, []byte{1})
}

// 发送前统一加锁处理 避免同时发送冲突
func (fun *Fun) getConnInfoAndSend(id string, callback func(loadConnInfo connInfoType) error) bool {
	if connInfo, ok := fun.connList.Load(id); ok {
		loadConnInfo := connInfo.(connInfoType)
		loadConnInfo.mu.Lock()
		err := callback(loadConnInfo)
		loadConnInfo.mu.Unlock()
		return err == nil
	}
	return false
}
