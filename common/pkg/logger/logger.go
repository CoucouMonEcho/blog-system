package logger

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// 单一职责：统一的全局日志中心 + 上下文获取能力

type Logger interface {
	Debug(format string, args ...any)
	Info(format string, args ...any)
	Warn(format string, args ...any)
	Error(format string, args ...any)
	Fatal(format string, args ...any)
	LogWithContext(service, path, level, format string, args ...any)
	LogWithContextAndRequestID(ctx context.Context, service, path, level, format string, args ...any)
}

type StructuredLogger struct {
	debugLogger *log.Logger
	infoLogger  *log.Logger
	warnLogger  *log.Logger
	errorLogger *log.Logger
	fatalLogger *log.Logger
	level       string
}

type LogLevel int

const (
	DEBUG LogLevel = iota
	INFO
	WARN
	ERROR
	FATAL
)

type Config struct {
	Level string `yaml:"level"`
	Path  string `yaml:"path"`
}

// 全局与上下文管理
var (
	gMu sync.RWMutex
	gL  Logger
)

type ctxKey struct{}

// Init 在服务启动时调用一次
func Init(l Logger) {
	gMu.Lock()
	gL = l
	gMu.Unlock()
}

// L 获取全局 Logger
func L() Logger {
	gMu.RLock()
	defer gMu.RUnlock()
	return gL
}

// With 将 Logger 写入上下文
func With(ctx context.Context, l Logger) context.Context {
	return context.WithValue(ctx, ctxKey{}, l)
}

// From 从上下文提取 Logger，若没有则返回全局
func From(ctx context.Context) Logger {
	if ctx != nil {
		if v := ctx.Value(ctxKey{}); v != nil {
			if l, ok := v.(Logger); ok {
				return l
			}
		}
	}
	return L()
}

// GenerateRequestID generates a random request ID
func GenerateRequestID() string {
	bytes := make([]byte, 8)
	_, _ = rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

// WithRequestID returns a new context with the request ID
func WithRequestID(ctx context.Context, requestID string) context.Context {
	return context.WithValue(ctx, "request_id", requestID)
}

// GetRequestID returns the request ID from context
func GetRequestID(ctx context.Context) string {
	if id, ok := ctx.Value("request_id").(string); ok {
		return id
	}
	return ""
}

// NewLogger 创建新的日志器
func NewLogger(cfg *Config) (Logger, error) {
	logDir := filepath.Dir(cfg.Path)
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return nil, fmt.Errorf("创建日志目录失败: %w", err)
	}
	logFile, err := os.OpenFile(cfg.Path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return nil, fmt.Errorf("创建日志文件失败: %w", err)
	}
	multiWriter := io.MultiWriter(os.Stdout, logFile)

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

func (l *StructuredLogger) shouldLog(level LogLevel) bool { return level >= l.getLogLevel() }

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
func (l *StructuredLogger) Fatal(format string, args ...any) {
	if l.shouldLog(FATAL) {
		l.fatalLogger.Printf(format, args...)
		os.Exit(1)
	}
}

func (l *StructuredLogger) LogWithContext(service, path, level, format string, args ...any) {
	timestamp := time.Now().Format("2006-01-02 15:04:05.000")
	prefix := fmt.Sprintf("[%s][%s][%s][%s] ", timestamp, service, path, level)
	switch level {
	case "DEBUG":
		if l.shouldLog(DEBUG) {
			l.debugLogger.Printf(prefix+format, args...)
		}
	case "INFO":
		if l.shouldLog(INFO) {
			l.infoLogger.Printf(prefix+format, args...)
		}
	case "WARN":
		if l.shouldLog(WARN) {
			l.warnLogger.Printf(prefix+format, args...)
		}
	case "ERROR":
		if l.shouldLog(ERROR) {
			l.errorLogger.Printf(prefix+format, args...)
		}
	case "FATAL":
		if l.shouldLog(FATAL) {
			l.fatalLogger.Printf(prefix+format, args...)
			os.Exit(1)
		}
	}
}

func (l *StructuredLogger) LogWithContextAndRequestID(ctx context.Context, service, path, level, format string, args ...any) {
	timestamp := time.Now().Format("2006-01-02 15:04:05.000")
	requestID := GetRequestID(ctx)
	var prefix string
	if requestID != "" {
		prefix = fmt.Sprintf("[%s][%s][%s][%s][%s] ", timestamp, service, path, level, requestID)
	} else {
		prefix = fmt.Sprintf("[%s][%s][%s][%s] ", timestamp, service, path, level)
	}
	switch level {
	case "DEBUG":
		if l.shouldLog(DEBUG) {
			l.debugLogger.Printf(prefix+format, args...)
		}
	case "INFO":
		if l.shouldLog(INFO) {
			l.infoLogger.Printf(prefix+format, args...)
		}
	case "WARN":
		if l.shouldLog(WARN) {
			l.warnLogger.Printf(prefix+format, args...)
		}
	case "ERROR":
		if l.shouldLog(ERROR) {
			l.errorLogger.Printf(prefix+format, args...)
		}
	case "FATAL":
		if l.shouldLog(FATAL) {
			l.fatalLogger.Printf(prefix+format, args...)
			os.Exit(1)
		}
	}
}
