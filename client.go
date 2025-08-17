package fun

import (
	"fmt"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

var clientId = "lGbk6IVcT965Qs_zb30KS"
var writeMutex sync.Mutex // 添加互斥锁
func SetTestClientId(id string) {
	clientId = id
}

var testClient *websocket.Conn = nil

func client(port uint16) {
	url := fmt.Sprintf("ws://127.0.0.1:%d?id=%s", port, clientId)
	conn, _, err := websocket.DefaultDialer.Dial(url, nil)
	for err != nil {
		conn, _, err = websocket.DefaultDialer.Dial(url, nil)
		time.Sleep(100 * time.Millisecond)
	}
	testClient = conn
	go func() {
		for {
			writeMutex.Lock() // 加锁
			err = conn.WriteMessage(websocket.BinaryMessage, []byte{0})
			writeMutex.Unlock()
			time.Sleep(5 * time.Second)
		}
	}()
	go func() {
		for {
			messageType, message, _ := conn.ReadMessage()
			if messageType == websocket.TextMessage {
				testMessageQueue <- message
			}
		}
	}()
}
