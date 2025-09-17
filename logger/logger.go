package logger

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"go.uber.org/zap/zapcore"

	"go.uber.org/zap"
)

var (
	Logger *zap.SugaredLogger
	root   string
)

// 初始化一个默认的logger
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

	baseLogger, err := NewLogger(logLevel, file, maxSize, maxBackups, maxAge)
	if err != nil {
		return err
	}

	Logger = baseLogger.Sugar()
	return nil
}

// NewLogger 创建日志实例
func NewLogger(logLevel, file string, maxSize, maxBackups, maxAge int) (*zap.Logger, error) {
	// 配置日志级别
	var level zapcore.Level
	switch logLevel {
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

	// 配置日志输出
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.TimeKey = "timestamp"
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder
	// 测试使用内置编码器
	encoderConfig.EncodeCaller = callerEncoder

	// 创建文件输出
	var cores []zapcore.Core

	// 控制台输出 - 使用自定义编码器
	consoleEncoder := zapcore.NewConsoleEncoder(encoderConfig)
	cores = append(cores, zapcore.NewCore(consoleEncoder, zapcore.AddSync(os.Stdout), level))

	// 文件输出（如果配置了文件路径）
	if file != "" {
		// 确保日志目录存在
		if err := os.MkdirAll("./logs", 0755); err != nil {
			return nil, err
		}

		file, err := os.OpenFile(file, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			return nil, err
		}

		// 文件使用JSON格式，但也使用自定义caller编码器
		fileEncoderConfig := encoderConfig
		fileEncoder := zapcore.NewJSONEncoder(fileEncoderConfig)
		cores = append(cores, zapcore.NewCore(fileEncoder, zapcore.AddSync(file), level))
	}

	// 组合所有core
	core := zapcore.NewTee(cores...)
	logger := zap.New(core, zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel))

	return logger, nil
}

// callerEncoder 使用工作目录计算相对路径
func callerEncoder(caller zapcore.EntryCaller, enc zapcore.PrimitiveArrayEncoder) {
	fullPath := caller.File

	// 如果设置了项目根路径，计算相对路径
	if root != "" {
		if relPath, err := filepath.Rel(root, fullPath); err == nil {
			// 统一使用正斜杠
			relPath = strings.ReplaceAll(relPath, "\\", "/")
			enc.AppendString(relPath + ":" + strconv.Itoa(caller.Line))
			return
		}
	}

	// 备选方案：使用默认格式
	enc.AppendString(caller.TrimmedPath())
}

// Debug 便捷的日志方法
func Debug(msg string, keysAndValues ...interface{}) {
	Logger.Desugar().WithOptions(zap.AddCallerSkip(1)).Sugar().Debugw(msg, keysAndValues...)
}

func Info(msg string, keysAndValues ...interface{}) {
	Logger.Desugar().WithOptions(zap.AddCallerSkip(1)).Sugar().Infow(msg, keysAndValues...)
}

func Warn(msg string, keysAndValues ...interface{}) {
	Logger.Desugar().WithOptions(zap.AddCallerSkip(1)).Sugar().Warnw(msg, keysAndValues...)
}

func Error(msg string, keysAndValues ...interface{}) {
	Logger.Desugar().WithOptions(zap.AddCallerSkip(1)).Sugar().Errorw(msg, keysAndValues...)
}

func Fatal(msg string, keysAndValues ...interface{}) {
	Logger.Desugar().WithOptions(zap.AddCallerSkip(1)).Sugar().Fatalw(msg, keysAndValues...)
}

func Fatalf(format string, keysAndValues ...interface{}) {
	Logger.Desugar().WithOptions(zap.AddCallerSkip(1)).Sugar().Fatalf(format, keysAndValues...)
}

func IsDebug() bool {
	return Logger.Desugar().Core().Enabled(zapcore.DebugLevel)
}

// Sync 刷新日志缓冲区
func Sync() {
	if Logger != nil {
		Logger.Sync()
	}
}
