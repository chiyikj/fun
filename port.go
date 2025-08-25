package fun

import (
	"fmt"
	"net"
	"strconv"
)

func randomPort(addr ...uint16) uint16 {
	var port uint16
	if len(addr) == 0 {
		port = 3000
	} else {
		port = addr[0]
	}
	l, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%d", port))
	for err != nil {
		port += 1
		l, err = net.Listen("tcp", fmt.Sprintf("0.0.0.0:%d", port))
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
		port = strconv.Itoa(int(randomPort(addr[0])))
	}
	return port
}
