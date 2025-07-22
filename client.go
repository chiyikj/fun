package fun

import (
	"fmt"
	"time"

	"github.com/gorilla/websocket"
)

func client(id string, port uint16) *websocket.Conn {
	url := fmt.Sprintf("ws://localhost:%d?id=%s", port, id)
	conn, _, err := websocket.DefaultDialer.Dial(url, nil)
	for err != nil {
		conn, _, err = websocket.DefaultDialer.Dial(url, nil)
		time.Sleep(100 * time.Millisecond)
	}
	return conn
}

type ClientInfo struct {
	Id string
}

func GetClientInfo(id string) ClientInfo {
	return ClientInfo{
		Id: id,
	}
}
