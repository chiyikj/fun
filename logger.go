package fun

import (
	"fmt"
	"runtime"
	"strings"
	"time"
)

const (
	TraceLevel = iota
	DebugLevel
	InfoLevel
	ErrorLevel
	WarnLevel
	PanicLevel
)

const (
	TerminalMode = iota
	FileMode
)

type Logger struct {
	Level          uint8
	Mode           uint8
	MaxSizeFile    uint8
	MaxNumberFiles uint8
	ExpireLogsDays uint8
}

func getCurrentTime() string {
	return time.Now().Format("2006-01-02 15:04:05")
}

func TraceLogger() {
	pc, _, _, _ := runtime.Caller(1)
	funcObj := runtime.FuncForPC(pc)
	a := []string{"(", "*", ")"}
	name := "[" + padString(strings.ReplaceAll(funcObj.Name(), "/", "."), 40, true) + "]"
	for _, v := range a {
		name = strings.ReplaceAll(name, v, "")
	}
	fmt.Println(padString("["+getCurrentTime()+"]", 5, true) + "[" + padString("DEBUG", 6, true) + "] " + name)
}

func DebugLogger() {

}

func InfoLogger() {

}

func ErrorLogger() {

}
func WarnLogger() {

}

func panicLogger() {

}

func padString(str string, totalLength int, leftAlign bool) string {
	if leftAlign {
		fmt.Println(fmt.Sprintf("%-*s", totalLength, str)[0:totalLength])
		return fmt.Sprintf("%-*s", totalLength, str)[0:totalLength] // 左对齐
	} else {
		return fmt.Sprintf("%*s", totalLength, str)[0:totalLength] // 右对齐
	}
}
