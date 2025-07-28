package main

import (
	"fun"
	"fun/abc/User"
	"runtime"
)

func main() {
	add()
	for {
	}
}

func add() {
	defer func() {
		if err := recover(); err != nil {
			if err1, ok := err.(error); ok {
				stackBuf := make([]byte, 8192)
				stackSize := runtime.Stack(stackBuf, false)
				stackTrace := string(stackBuf[:stackSize])
				fun.PanicLogger(err1.Error() + "\n" + stackTrace)
			}
		}
	}()
	fun.TraceLogger("111")
	fun.InfoLogger("111")
	a := User.Name{}
	a.Name()
}
