package logger

import (
	"io"
	"log"
	"os"
	"path/filepath"
	"sync"
)

// Logger 全局日志
type Logger interface {
	Debug(format string, args ...any)
	Info(format string, args ...any)
	Warn(format string, args ...any)
	Error(format string, args ...any)
}

type LogLevel int

const (
	DEBUG LogLevel = iota
	INFO
	WARN
	ERROR
)

type Config struct {
	Level string `yaml:"level"`
	Path  string `yaml:"path"`
}

var (
	mutex    sync.RWMutex
	instance Logger
)

type StructuredLogger struct {
	debugLogger *log.Logger
	infoLogger  *log.Logger
	warnLogger  *log.Logger
	errorLogger *log.Logger

	level string
}

// Log 获取全局 Logger
func Log() Logger {

	mutex.RLock()
	if instance != nil {
		defer mutex.RUnlock()
		return instance
	}
	mutex.RUnlock()

	debugLogger := log.New(os.Stdout, "[DEBUG] ", log.Ldate|log.Ltime|log.Lmicroseconds)
	infoLogger := log.New(os.Stdout, "[INFO] ", log.Ldate|log.Ltime|log.Lmicroseconds)
	warnLogger := log.New(os.Stdout, "[WARN] ", log.Ldate|log.Ltime|log.Lmicroseconds)
	errorLogger := log.New(os.Stdout, "[ERROR] ", log.Ldate|log.Ltime|log.Lmicroseconds)

	mutex.Lock()
	if instance == nil {
		instance = &StructuredLogger{
			debugLogger: debugLogger,
			infoLogger:  infoLogger,
			warnLogger:  warnLogger,
			errorLogger: errorLogger,
			level:       "info",
		}
	}
	l := instance
	mutex.Unlock()
	return l
}

// Init 初始化全局日志处理器
func Init(cfg *Config) {
	mutex.Lock()
	defer mutex.Unlock()
	if instance != nil {
		return
	}

	logDir := filepath.Dir(cfg.Path)
	if err := os.MkdirAll(logDir, 0755); err != nil {
		log.Fatalf("创建日志目录失败: %v", err)
		return
	}
	logFile, err := os.OpenFile(cfg.Path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("创建日志文件失败: %v", err)
		return
	}
	multiWriter := io.MultiWriter(os.Stdout, logFile)

	debugLogger := log.New(multiWriter, "[DEBUG] ", log.Ldate|log.Ltime|log.Lmicroseconds)
	infoLogger := log.New(multiWriter, "[INFO] ", log.Ldate|log.Ltime|log.Lmicroseconds)
	warnLogger := log.New(multiWriter, "[WARN] ", log.Ldate|log.Ltime|log.Lmicroseconds)
	errorLogger := log.New(multiWriter, "[ERROR] ", log.Ldate|log.Ltime|log.Lmicroseconds)

	instance = &StructuredLogger{
		debugLogger: debugLogger,
		infoLogger:  infoLogger,
		warnLogger:  warnLogger,
		errorLogger: errorLogger,
		level:       cfg.Level,
	}
}

func (l *StructuredLogger) Debug(format string, args ...any) {
	if l.shouldLog(DEBUG) {
		l.debugLogger.Printf(format, args...)
	}
}

func (l *StructuredLogger) Info(format string, args ...any) {
	if l.shouldLog(INFO) {
		l.infoLogger.Printf(format, args...)
	}
}

func (l *StructuredLogger) Warn(format string, args ...any) {
	if l.shouldLog(WARN) {
		l.warnLogger.Printf(format, args...)
	}
}

func (l *StructuredLogger) Error(format string, args ...any) {
	if l.shouldLog(ERROR) {
		l.errorLogger.Printf(format, args...)
	}
}

func (l *StructuredLogger) shouldLog(level LogLevel) bool {
	return level >= l.getLogLevel()
}

func (l *StructuredLogger) getLogLevel() LogLevel {
	switch l.level {
	case "debug":
		return DEBUG
	case "info":
		return INFO
	case "warn":
		return WARN
	case "error":
		return ERROR
	default:
		return INFO
	}
}
