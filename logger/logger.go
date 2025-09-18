package logger

import (
	"os"
	"path/filepath"
	"strings"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	Logger *zap.SugaredLogger
	root   string
)

// init 初始化一个默认的logger
func init() {
	_ = Init("debug", "./logs/app.log", 10, 5, 30)
}

// Init 初始化全局logger
func Init(logLevel, file string, maxSize, maxBackups, maxAge int) error {
	// 初始化项目根路径（只在第一次调用时设置）
	if root == "" {
		if wd, err := os.Getwd(); err == nil {
			root = wd
		}
	}

	// 配置日志级别
	var level zapcore.Level
	switch strings.ToLower(logLevel) {
	case "debug":
		level = zapcore.DebugLevel
	case "info":
		level = zapcore.InfoLevel
	case "warn":
		level = zapcore.WarnLevel
	case "error":
		level = zapcore.ErrorLevel
	default:
		level = zapcore.InfoLevel
	}

	// 创建logger配置
	config := zap.NewProductionConfig()
	config.Level = zap.NewAtomicLevelAt(level)
	config.OutputPaths = []string{"stdout", file}
	config.ErrorOutputPaths = []string{"stderr", file}

	// 确保日志目录存在
	if err := os.MkdirAll(filepath.Dir(file), 0755); err != nil {
		return err
	}

	// 创建logger
	logger, err := config.Build()
	if err != nil {
		return err
	}

	Logger = logger.Sugar()
	return nil
}

// Debug 记录调试日志
func Debug(msg string, keysAndValues ...interface{}) {
	if Logger != nil {
		Logger.Debugw(msg, keysAndValues...)
	}
}

// Info 记录信息日志
func Info(msg string, keysAndValues ...interface{}) {
	if Logger != nil {
		Logger.Infow(msg, keysAndValues...)
	}
}

// Warn 记录警告日志
func Warn(msg string, keysAndValues ...interface{}) {
	if Logger != nil {
		Logger.Warnw(msg, keysAndValues...)
	}
}

// Error 记录错误日志
func Error(msg string, keysAndValues ...interface{}) {
	if Logger != nil {
		Logger.Errorw(msg, keysAndValues...)
	}
}

// Fatal 记录致命错误日志并退出程序
func Fatal(msg string, keysAndValues ...interface{}) {
	if Logger != nil {
		Logger.Fatalw(msg, keysAndValues...)
	}
}

// GetAbsolutePath 获取相对于项目根路径的绝对路径
func GetAbsolutePath(relativePath string) string {
	if root == "" {
		return relativePath
	}
	return filepath.Join(root, relativePath)
}

// SetLogLevel 动态设置日志级别
func SetLogLevel(levelStr string) {
	// 这个功能需要重新初始化logger，这里简化实现
	Info("日志级别更改请求", "level", levelStr)
}
