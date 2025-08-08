package mock

import (
	"fmt"
	"net"
	"strconv"
)

func randomPort() uint16 {
	var port uint16 = 3000
	l, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", port))
	if err != nil {
		port = port + 1
		l, err = net.Listen("tcp", fmt.Sprintf(":%d", port))
	}
	defer func(l net.Listener) {
		_ = l.Close()
	}(l)
	return port
}

func isPort(addr []uint16) string {
	var port string
	if len(addr) == 0 {
		port = strconv.Itoa(int(randomPort()))
	} else {
		port = strconv.Itoa(int(addr[0]))
	}
	return port
}
