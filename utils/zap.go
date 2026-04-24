package utils

import (
	"os"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

var Logger *zap.Logger

// InitZapLogger 初始化所有日志（系统日志）
func InitZapLogger() {
	_ = os.MkdirAll("./logs", 0755)

	Logger = initLogger("logs/app.log")
}

func initLogger(filename string) *zap.Logger {
	encoder := getEncoder()
	writeSyncer := getLogWriter(filename)
	core := zapcore.NewCore(encoder, writeSyncer, zap.DebugLevel)

	return zap.New(core, zap.AddCaller())
}

func getEncoder() zapcore.Encoder {
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.EncodeTime = func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
		enc.AppendString(t.Format("2006-01-02 15:04:05"))
	}
	encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder
	return zapcore.NewConsoleEncoder(encoderConfig)
}

func getLogWriter(filename string) zapcore.WriteSyncer {
	lumber := &lumberjack.Logger{
		Filename:   filename,
		MaxSize:    100, // 单个文件最大 100M
		MaxBackups: 7,   // 保留 7 个备份
		MaxAge:     7,   // 保留 7 天
		Compress:   false,
		LocalTime:  true,
	}

	return zapcore.NewMultiWriteSyncer(
		zapcore.AddSync(lumber),
		zapcore.AddSync(os.Stdout),
	)
}
