package fun

import (
	"reflect"

	"github.com/gorilla/websocket"
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
	on, ok := loadConnInfo.onList.Load(requestId)
	if ok {
		method := fun.serviceList[on.(onType).serviceName].methodList[on.(onType).methodName]
		if method.method.Type.Out(0).Elem() == reflect.TypeOf(data) {
			result := success(data)
			result.Id = requestId
			map1 := ToLowerMap(result)
			err := loadConnInfo.conn.WriteJSON(map1)
			if err != nil {
				return false
			}
		} else {
			return false
		}
	} else {
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
