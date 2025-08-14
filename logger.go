package fun

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
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
	MaxSizeFile    uint8  //文件最大大小
	MaxNumberFiles uint64 //文件最多数量
	ExpireLogsDays uint8  //文件保留时间
}

const logFile = "../log/"

var logger *Logger = &Logger{
	Level:          TraceLevel,
	Mode:           TerminalMode,
	MaxSizeFile:    0,
	MaxNumberFiles: 0,
	ExpireLogsDays: 0,
}

type fileName struct {
	LoggerTime int64
	index      int32
}

// 日志通道

// 初始化日志系统
func init() {
	go deleteLogWorker() // 清理
}

func deleteLogWorker() {
	// 定期清理过期日志文件
	ticker := time.NewTicker(24 * time.Hour)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			if logger.Mode == FileMode {
				cleanupExpiredLogs()
			}
		}
	}
}

func getFileNameInfo(name string) fileName {
	// 分割文件名获取日期和索引部分
	fileNameParts := strings.Split(name, ".log.")

	// 检查文件名格式是否正确
	if len(fileNameParts) != 2 {
		WarnLogger(fmt.Sprintf("Invalid log filename format: %s", name))
		deleteLog(name)
		return fileName{}
	}

	// 解析日期部分
	dateLayout := "2006-01-02"
	dateString := fileNameParts[0]
	fileDate, err := time.Parse(dateLayout, dateString)
	if err != nil {
		WarnLogger(fmt.Sprintf("Failed to parse date from filename %s: %v", name, err))
		deleteLog(name)
		return fileName{}
	}

	// 解析文件索引
	indexString := fileNameParts[1]
	// 移除可能的 .log 扩展名
	indexString = strings.TrimSuffix(indexString, ".log")

	fileIndex, err := strconv.ParseInt(indexString, 10, 32)
	if err != nil {
		WarnLogger(fmt.Sprintf("Failed to parse index from filename %s: %v", name, err))
		deleteLog(name)
		return fileName{}
	}

	return fileName{
		index:      int32(fileIndex),
		LoggerTime: fileDate.UnixMilli(),
	}
}

func deleteLog(name string) {
	// 删除文件
	fullPath := filepath.Join(logFile, name)
	err := os.Remove(fullPath)
	if err != nil {
		if !os.IsNotExist(err) {
			ErrorLogger(fmt.Sprintf("Failed to delete log file %s: %v", fullPath, err))
		}
	} else {
		InfoLogger(fmt.Sprintf("Successfully deleted log file: %s", fullPath))
	}
}

func cleanupExpiredLogs() {
	// 如果没有设置过期天数，不进行清理
	if logger.ExpireLogsDays <= 0 {
		return
	}

	// 检查日志目录是否存在
	_, err := os.Stat(logFile)
	if os.IsNotExist(err) {
		return
	}

	if err != nil {
		ErrorLogger(fmt.Sprintf("Failed to check log directory status: %v", err))
		return
	}

	// 读取目录中的所有条目
	entries, err := os.ReadDir(logFile)
	if err != nil {
		ErrorLogger(fmt.Sprintf("Failed to read log directory: %v", err))
		return
	}

	// 计算过期时间阈值
	expireDuration := time.Duration(logger.ExpireLogsDays) * 24 * time.Hour
	currentTimeMillis := time.Now().UnixMilli()
	expireThreshold := currentTimeMillis - expireDuration.Milliseconds()

	// 遍历所有文件
	cleanedCount := 0
	for _, entry := range entries {
		if !entry.IsDir() { // 只处理文件
			fileNameInfo := getFileNameInfo(entry.Name())

			// 检查是否是有效的日志文件
			if fileNameInfo.LoggerTime == 0 {
				// 无效的日志文件名格式，跳过处理
				continue
			}

			// 检查文件是否过期
			if fileNameInfo.LoggerTime < expireThreshold {
				fullPath := filepath.Join(logFile, entry.Name())
				err := os.Remove(fullPath)
				if err != nil {
					if !os.IsNotExist(err) {
						ErrorLogger(fmt.Sprintf("Failed to delete expired log file %s: %v", fullPath, err))
					}
				} else {
					InfoLogger(fmt.Sprintf("Deleted expired log file: %s (date: %s, index: %d)",
						fullPath,
						time.Unix(0, fileNameInfo.LoggerTime*int64(time.Millisecond)).Format("2006-01-02"),
						fileNameInfo.index))
					cleanedCount++
				}
			}
		}
	}
	InfoLogger(fmt.Sprintf("Log cleanup completed. Deleted %d expired files", cleanedCount))
}

func fileLogger(text string) {
	// 确保日志目录存在
	currentDate := getCurrentData()

	_, err := os.Stat(logFile)
	if os.IsNotExist(err) {
		_ = os.MkdirAll(logFile, os.ModePerm)
		InfoLogger(fmt.Sprintf("Created log directory: %s", logFile))
	}

	// 构建日志文件路径
	logFileName := currentDate + ".log"
	logFilePath := filepath.Join(logFile, logFileName)

	// 获取当前目录下所有日志文件以确定下一个文件索引
	logFilePath = getNextLogFile(logFile, currentDate, text)
	// 写入日志文件
	file, err := os.OpenFile(logFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		ErrorLogger(fmt.Sprintf("Failed to open log file %s: %v", logFilePath, err))
		return
	}
	defer func(file *os.File) {
		_ = file.Close()
	}(file)

	_, err = file.WriteString(text + "\n")
	if err != nil {
		ErrorLogger(fmt.Sprintf("Failed to write to log file %s: %v", logFilePath, err))
	}
}

func removeOldestLogFile(entries []os.DirEntry) {
	if logger.MaxNumberFiles == 0 {
		return
	}

	if uint64(len(entries)) < logger.MaxNumberFiles {
		return
	}

	var newEntries []fileName
	for _, v := range entries {
		fileNameInfo := getFileNameInfo(v.Name())
		if fileNameInfo.LoggerTime != 0 {
			newEntries = append(newEntries, fileNameInfo)
		}
	}

	if uint64(len(newEntries)) < logger.MaxNumberFiles {
		return
	}

	delNum := uint64(len(newEntries)) - logger.MaxNumberFiles + 1
	sort.Slice(newEntries, func(i, j int) bool {
		if newEntries[i].LoggerTime != newEntries[j].LoggerTime {
			return newEntries[i].LoggerTime < newEntries[j].LoggerTime
		}
		// 日期相同时按索引比较
		return newEntries[i].index < newEntries[j].index
	})

	for i := 0; i < int(delNum); i++ {
		fileName := newEntries[i]
		t := time.Unix(0, fileName.LoggerTime*int64(time.Millisecond))
		fullPath := filepath.Join(logFile, t.Format("2006-01-02"))
		fileNamePath := filepath.Join(fullPath, t.Format("2006-01-02")+".log."+strconv.Itoa(int(fileName.index)))
		deleteLog(fileNamePath)
	}
}

// 获取下一个应该写入的日志文件
func getNextLogFile(dirPath, dateStr string, text string) string {
	// 读取目录中的所有文件
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		WarnLogger(fmt.Sprintf("Failed to read log directory %s: %v", dirPath, err))
		return filepath.Join(dirPath, dateStr+".log.1")
	}

	var maxIndex int32 = 0
	var existingFiles []string

	// 查找现有的日志文件
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasPrefix(entry.Name(), dateStr+".log") {
			existingFiles = append(existingFiles, entry.Name())
		}
	}

	// 如果没有现有文件，创建第一个
	if len(existingFiles) == 0 {
		removeOldestLogFile(entries)
		return filepath.Join(dirPath, dateStr+".log.1")
	}

	// 查找最大索引
	for _, fileName := range existingFiles {
		fileNameInfo := getFileNameInfo(fileName)
		if fileNameInfo.LoggerTime != 0 {
			if fileNameInfo.index > maxIndex {
				maxIndex = fileNameInfo.index
			}
		}
	}
	if maxIndex == 0 {
		removeOldestLogFile(entries)
		return filepath.Join(dirPath, dateStr+".log.1")
	}

	// 检查最大索引的文件是否超过大小限制
	if logger.MaxSizeFile > 0 && maxIndex > 0 {
		currentFile := filepath.Join(dirPath, fmt.Sprintf("%s.%d.log", dateStr, maxIndex))
		if fileInfo, err := os.Stat(currentFile); err == nil {
			maxSizeBytes := int64(logger.MaxSizeFile) * 1024 * 1024 // 转换为字节
			if fileInfo.Size()+int64(len(text)) > maxSizeBytes {
				removeOldestLogFile(entries)
				return filepath.Join(dirPath, fmt.Sprintf("%s.log.%d", dateStr, maxIndex+1))
			}
		}
	}

	// 使用当前最大索引文件
	return filepath.Join(dirPath, fmt.Sprintf("%s.log.%d", dateStr, maxIndex))
}

func ConfigLogger(log *Logger) {
	// 启动日志处理
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

func sendLogWorker(level uint8, message []any) {
	if logger.Level >= level {
		var text1 string
		for _, m := range message {
			var msgStr string
			var temp interface{}
			var trimmedStr string
			switch v := m.(type) {
			case string:
				// 处理字符串类型
				trimmedStr = strings.TrimSpace(v)
				if trimmedStr == "" {
					msgStr = fmt.Sprintf("%s", v)
					break
				}
				err := json.Unmarshal([]byte(trimmedStr), &temp)
				if err != nil {
					// 不是有效的JSON字符串，直接输出
					msgStr = fmt.Sprintf("%s", v)
					break
				}
				// 检查解析后的数据是否为对象或数组

			default:
				// 处理非字符串类型
				bs, _ := json.Marshal(v)
				err := json.Unmarshal(bs, &temp)
				if err != nil {
					msgStr = fmt.Sprintf("%v", v)
					break
				}
				trimmedStr = string(bs)
			}
			switch temp.(type) {
			case map[string]interface{}:
				// 是JSON对象
				var out bytes.Buffer
				json.Indent(&out, []byte(trimmedStr), "", "\t")
				msgStr = fmt.Sprintf("\n%s", out.String())
			case []interface{}:
				// 是JSON数组
				var out bytes.Buffer
				json.Indent(&out, []byte(trimmedStr), "", "\t")
				msgStr = fmt.Sprintf("\n%s", out.String())
			default:
				// 是字符串、数值、布尔值或null，直接输出
				msgStr = fmt.Sprintf("%v", m)
			}
			text1 += msgStr + " "
		}
		text := "[" + getCurrentTime() + "] [" + padString(getLevelName(level), 7) + "] " + getMethodNameLogger() + text1
		if logger.Mode == FileMode {
			// 文件模式
			fileLogger(text)
		} else {
			fmt.Println(text)
		}
	}
}

func DebugLogger(message ...any) {
	sendLogWorker(DebugLevel, message)
}

func InfoLogger(message ...any) {
	sendLogWorker(InfoLevel, message)
}

func TraceLogger(message ...any) {
	sendLogWorker(TraceLevel, message)
}

func ErrorLogger(message ...any) {
	sendLogWorker(ErrorLevel, message)
}
func WarnLogger(message ...any) {
	sendLogWorker(WarnLevel, message)
}

func PanicLogger(message ...any) {
	sendLogWorker(PanicLevel, message)
}

func padString(str string, totalLength int) string {
	return fmt.Sprintf("%-*s", totalLength, str)[0:totalLength] // 左对齐
}
