package logger

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"time"
)

// Logger 结构化日志接口
type Logger interface {
	Debug(format string, args ...any)
	Info(format string, args ...any)
	Warn(format string, args ...any)
	Error(format string, args ...any)
	Fatal(format string, args ...any)
	LogWithContext(service, path, level, format string, args ...any)
}

// StructuredLogger 结构化日志实现
type StructuredLogger struct {
	debugLogger *log.Logger
	infoLogger  *log.Logger
	warnLogger  *log.Logger
	errorLogger *log.Logger
	fatalLogger *log.Logger
	level       string
}

// LogLevel 日志级别
type LogLevel int

const (
	DEBUG LogLevel = iota
	INFO
	WARN
	ERROR
	FATAL
)

// Config 日志配置
type Config struct {
	Level string `yaml:"level"`
	Path  string `yaml:"path"`
}

// NewLogger 创建新的日志器
func NewLogger(cfg *Config) (Logger, error) {
	// 确保日志目录存在
	logDir := filepath.Dir(cfg.Path)
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return nil, fmt.Errorf("创建日志目录失败: %w", err)
	}

	// 创建日志文件
	logFile, err := os.OpenFile(cfg.Path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return nil, fmt.Errorf("创建日志文件失败: %w", err)
	}

	// 同时输出到文件和控制台
	multiWriter := io.MultiWriter(os.Stdout, logFile)

	// 创建不同级别的日志器
	debugLogger := log.New(multiWriter, "[DEBUG] ", log.Ldate|log.Ltime|log.Lmicroseconds)
	infoLogger := log.New(multiWriter, "[INFO] ", log.Ldate|log.Ltime|log.Lmicroseconds)
	warnLogger := log.New(multiWriter, "[WARN] ", log.Ldate|log.Ltime|log.Lmicroseconds)
	errorLogger := log.New(multiWriter, "[ERROR] ", log.Ldate|log.Ltime|log.Lmicroseconds)
	fatalLogger := log.New(multiWriter, "[FATAL] ", log.Ldate|log.Ltime|log.Lmicroseconds)

	return &StructuredLogger{
		debugLogger: debugLogger,
		infoLogger:  infoLogger,
		warnLogger:  warnLogger,
		errorLogger: errorLogger,
		fatalLogger: fatalLogger,
		level:       cfg.Level,
	}, nil
}

// getLogLevel 获取日志级别
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
	case "fatal":
		return FATAL
	default:
		return INFO
	}
}

// shouldLog 判断是否应该记录日志
func (l *StructuredLogger) shouldLog(level LogLevel) bool {
	return level >= l.getLogLevel()
}

// Debug 调试日志
func (l *StructuredLogger) Debug(format string, args ...any) {
	if l.shouldLog(DEBUG) {
		l.debugLogger.Printf(format, args...)
	}
}

// Info 信息日志
func (l *StructuredLogger) Info(format string, args ...any) {
	if l.shouldLog(INFO) {
		l.infoLogger.Printf(format, args...)
	}
}

// Warn 警告日志
func (l *StructuredLogger) Warn(format string, args ...any) {
	if l.shouldLog(WARN) {
		l.warnLogger.Printf(format, args...)
	}
}

// Error 错误日志
func (l *StructuredLogger) Error(format string, args ...any) {
	if l.shouldLog(ERROR) {
		l.errorLogger.Printf(format, args...)
	}
}

// Fatal 致命错误日志
func (l *StructuredLogger) Fatal(format string, args ...any) {
	if l.shouldLog(FATAL) {
		l.fatalLogger.Printf(format, args...)
		os.Exit(1)
	}
}

// LogWithContext 带上下文的日志
func (l *StructuredLogger) LogWithContext(service, path, level, format string, args ...any) {
	timestamp := time.Now().Format("2006-01-02 15:04:05.000")
	context := fmt.Sprintf("[%s][%s][%s][%s] ", timestamp, service, path, level)

	switch level {
	case "DEBUG":
		if l.shouldLog(DEBUG) {
			l.debugLogger.Printf(context+format, args...)
		}
	case "INFO":
		if l.shouldLog(INFO) {
			l.infoLogger.Printf(context+format, args...)
		}
	case "WARN":
		if l.shouldLog(WARN) {
			l.warnLogger.Printf(context+format, args...)
		}
	case "ERROR":
		if l.shouldLog(ERROR) {
			l.errorLogger.Printf(context+format, args...)
		}
	case "FATAL":
		if l.shouldLog(FATAL) {
			l.fatalLogger.Printf(context+format, args...)
			os.Exit(1)
		}
	}
}
