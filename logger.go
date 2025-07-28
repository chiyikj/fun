package fun

import (
	"fmt"
	"os"
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
	MaxSizeFile    uint8 //文件最大大小
	MaxNumberFiles uint8 //文件最多数量
	ExpireLogsDays uint8 //文件保留时间
}

// 日志消息结构体
type logMessage struct {
	level   uint8
	message string
}

const logFile = "./log"

var logger *Logger = &Logger{
	Level:          TraceLevel,
	Mode:           FileMode,
	MaxSizeFile:    0,
	MaxNumberFiles: 0,
	ExpireLogsDays: 0,
}

// 日志通道
var logChan chan logMessage

// 初始化日志系统
func init() {
	defer func() {
		if err := recover(); err != nil {
			stackBuf := make([]byte, 8192)
			stackSize := runtime.Stack(stackBuf, false)
			stackTrace := string(stackBuf[:stackSize])
			PanicLogger(getErrorString(err) + "\n" + stackTrace)
		}
	}()
	logChan = make(chan logMessage, 1000) // 创建带缓冲的通道
	go logWorker()
	go deleteLogWorker() // 清理
}

func deleteLogWorker() {
	defer func() {
		if err := recover(); err != nil {
			stackBuf := make([]byte, 8192)
			stackSize := runtime.Stack(stackBuf, false)
			stackTrace := string(stackBuf[:stackSize])
			PanicLogger(getErrorString(err) + "\n" + stackTrace)
		}
	}()

}

func logWorker() {
	defer func() {
		if err := recover(); err != nil {
			stackBuf := make([]byte, 8192)
			stackSize := runtime.Stack(stackBuf, false)
			stackTrace := string(stackBuf[:stackSize])
			PanicLogger(getErrorString(err) + "\n" + stackTrace)
		}
	}()
	for msg := range logChan {
		text := "[" + getCurrentTime() + "] [" + padString(getLevelName(msg.level), 7) + "] " + msg.message
		if logger.Mode == FileMode {
			// 文件模式
			fileLogger(text)
		} else {
			fmt.Println(text)
		}
	}
}

func fileLogger(text string) {
	fullPath := logFile + "/" + getCurrentData()
	_, err := os.Stat(fullPath)
	if os.IsNotExist(err) {
		err = os.MkdirAll(fullPath, os.ModePerm)
		if err != nil {
			panic(err)
		}
	}

}

func ConfigLogger(log *Logger) {
	// 启动日志处理协程
	logger = log
}

func getCurrentTime() string {
	return time.Now().Format("2006-01-02 15:04:05")
}

func getCurrentData() string {
	return time.Now().Format("2006-01-02")
}

func getMethodNameLogger() string {
	pc, _, _, _ := runtime.Caller(3)
	fn := runtime.FuncForPC(pc)
	// 定义需要移除的字符
	charsToRemove := []string{"(", "*", ")"}
	name := fn.Name()
	for _, char := range charsToRemove {
		name = strings.ReplaceAll(name, char, "")
	}
	funcName := "[" + padString(strings.ReplaceAll(name, "/", "."), 40) + "] "

	return funcName
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

func sendLogWorker(level uint8, message string) {
	if logger.Level >= level {
		logChan <- logMessage{
			level:   level,
			message: getMethodNameLogger() + message,
		}
	}
}

func DebugLogger(message string) {
	sendLogWorker(DebugLevel, message)
}

func InfoLogger(message string) {
	sendLogWorker(InfoLevel, message)
}

func TraceLogger(message string) {
	sendLogWorker(TraceLevel, message)
}

func ErrorLogger(message string) {
	sendLogWorker(ErrorLevel, message)
}
func WarnLogger(message string) {
	sendLogWorker(WarnLevel, message)
}

func PanicLogger(message string) {
	sendLogWorker(PanicLevel, message)
}

func padString(str string, totalLength int) string {
	return fmt.Sprintf("%-*s", totalLength, str)[0:totalLength] // 左对齐
}
