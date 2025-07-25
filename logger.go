package fun

import (
	"fmt"
	"runtime"
	"strings"
	"time"
)

const (
	PanicLevel = iota
	ErrorLevel
	WarnLevel
	InfoLevel
	DebugLevel
	TraceLevel
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

// 日志消息结构体
type logMessage struct {
	level   uint8
	message string
}

// 日志通道
var logChan chan logMessage

// 初始化日志系统
func init() {
	logChan = make(chan logMessage, 1000) // 创建带缓冲的通道
	go logWorker()                        // 启动日志处理协程
}

func logWorker() {
	for msg := range logChan {
		// 统一处理所有日志消息
		fmt.Println("[" + getCurrentTime() + "] [" + getLevelName(msg.level) + "] " + msg.message)
	}
}

func getCurrentTime() string {
	return time.Now().Format("2006-01-02 15:04:05")
}

func getMethodNameLogger() string {
	pc, _, _, _ := runtime.Caller(3)
	funcObj := runtime.FuncForPC(pc)
	a := []string{"(", "*", ")"}
	name := "[" + padString(strings.ReplaceAll(funcObj.Name(), "/", "."), 40, true) + "]"
	for _, v := range a {
		name = strings.ReplaceAll(name, v, "")
	}
	return name
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

func getLevelName(level uint8) string {
	switch level {
	case TraceLevel:
		return "TRACE"
	case DebugLevel:
		return "DEBUG"
	case InfoLevel:
		return "INFO"
	case ErrorLevel:
		return "ERROR"
	case WarnLevel:
		return "WARN"
	default:
		return "PANIC"
	}
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
		return fmt.Sprintf("%-*s", totalLength, str)[0:totalLength] // 左对齐
	} else {
		return fmt.Sprintf("%*s", totalLength, str)[0:totalLength] // 右对齐
	}
}
