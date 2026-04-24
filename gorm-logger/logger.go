package gorm_logger

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"time"

	gormlogger "gorm.io/gorm/logger"
	"gorm.io/gorm/utils"
)

const (
	Reset       = "\033[0m"
	Red         = "\033[31m"
	Green       = "\033[32m"
	Yellow      = "\033[33m"
	Blue        = "\033[34m"
	Magenta     = "\033[35m"
	Cyan        = "\033[36m"
	White       = "\033[37m"
	BlueBold    = "\033[34;1m"
	MagentaBold = "\033[35;1m"
	RedBold     = "\033[31;1m"
	YellowBold  = "\033[33;1m"
)

const (
	// Silent silent log level
	Silent gormlogger.LogLevel = iota + 1
	// Error error log level
	Error
	// Warn warn log level
	Warn
	// Info info log level
	Info
)

type FileLogConfig struct {
	gormlogger.Config
}
type PathOption struct {
	Path string
	Name string
}

type StdFileLogger struct {
	FileLogConfig
	infoStr, warnStr, errStr            string
	traceStr, traceErrStr, traceWarnStr string
	path, name                          string
}

func NewLogger(config FileLogConfig, path PathOption) *StdFileLogger {
	var (
		infoStr      = "%s\n[info] "
		warnStr      = "%s\n[warn] "
		errStr       = "%s\n[error] "
		traceStr     = "%s\n[%.3fms] [rows:%v] %s"
		traceWarnStr = "%s %s\n[%.3fms] [rows:%v] %s"
		traceErrStr  = "%s %s\n[%.3fms] [rows:%v] %s"
	)

	if config.Colorful {
		infoStr = Green + "%s\n" + Reset + Green + "[info] " + Reset
		warnStr = BlueBold + "%s\n" + Reset + Magenta + "[warn] " + Reset
		errStr = Magenta + "%s\n" + Reset + Red + "[error] " + Reset
		traceStr = Green + "%s\n" + Reset + Yellow + "[%.3fms] " + BlueBold + "[rows:%v]" + Reset + " %s"
		traceWarnStr = Green + "%s " + Yellow + "%s\n" + Reset + RedBold + "[%.3fms] " + Yellow + "[rows:%v]" + Magenta + " %s" + Reset
		traceErrStr = RedBold + "%s " + MagentaBold + "%s\n" + Reset + Yellow + "[%.3fms] " + BlueBold + "[rows:%v]" + Reset + " %s"
	}
	if path.Path == "" {
		dir, _ := os.Getwd()
		path.Path = dir
	}
	if path.Name == "" {
		panic("文件名不能为空")
	}
	return &StdFileLogger{
		FileLogConfig: config,
		path:          path.Path,
		name:          path.Name,
		infoStr:       infoStr,
		warnStr:       warnStr,
		errStr:        errStr,
		traceStr:      traceStr,
		traceWarnStr:  traceWarnStr,
		traceErrStr:   traceErrStr,
	}
}

func (logger *StdFileLogger) printf(msg string, data ...interface{}) {
	//得到要打印的内容
	content := fmt.Sprintf(msg, data...)
	//得到文件名:
	now := time.Now()
	filePath := filepath.Join(logger.path, logger.name)

	// 格式化时间
	formatted := now.Format("2006-01-02 15:04:05")
	content = formatted + " " + content
	//保存到文件
	logger.LogToFile(filePath, content+"\n\r", false)
}

func (logger *StdFileLogger) LogToFile(filename, msg string, close bool) {
	// 输出到文件
	file, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Println("日志文件的打开错误 :", err)
	}
	if close {
		defer file.Close()
	}
	if _, err := file.WriteString(msg); err != nil {
		fmt.Println("写入日志文件错误 :", err)
	}
}

func (logger *StdFileLogger) Trace(ctx context.Context, begin time.Time, fc func() (sql string, rowsAffected int64), err error) {
	if logger.LogLevel <= Silent {
		return
	}

	elapsed := time.Since(begin)
	switch {
	case err != nil && logger.LogLevel >= Error && (!errors.Is(err, gormlogger.ErrRecordNotFound) || !logger.IgnoreRecordNotFoundError):
		sql, rows := fc()
		//得到stack
		stack := logger.PrintStackTrace(err)
		if rows == -1 {
			logger.printf(logger.traceErrStr+"\n"+stack, utils.FileWithLineNum(), err, float64(elapsed.Nanoseconds())/1e6, "-", sql)
		} else {
			logger.printf(logger.traceErrStr+"\n"+stack, utils.FileWithLineNum(), err, float64(elapsed.Nanoseconds())/1e6, rows, sql)
		}
	case elapsed > logger.SlowThreshold && logger.SlowThreshold != 0 && logger.LogLevel >= Warn:
		sql, rows := fc()
		slowLog := fmt.Sprintf("SLOW SQL >= %v", logger.SlowThreshold)
		if rows == -1 {
			logger.printf(logger.traceWarnStr, utils.FileWithLineNum(), slowLog, float64(elapsed.Nanoseconds())/1e6, "-", sql)
		} else {
			logger.printf(logger.traceWarnStr, utils.FileWithLineNum(), slowLog, float64(elapsed.Nanoseconds())/1e6, rows, sql)
		}
	case logger.LogLevel == Info:
		sql, rows := fc()
		if rows == -1 {
			logger.printf(logger.traceStr, utils.FileWithLineNum(), float64(elapsed.Nanoseconds())/1e6, "-", sql)
		} else {
			logger.printf(logger.traceStr, utils.FileWithLineNum(), float64(elapsed.Nanoseconds())/1e6, rows, sql)
		}
	}
}

func (logger *StdFileLogger) PrintStackTrace(err error) string {
	buf := bytes.NewBuffer(nil)

	for i := 0; ; i++ {
		pc, file, line, ok := runtime.Caller(i)
		if !ok {
			break
		}

		fmt.Fprintf(buf, "%d: %s:%d (0x%x)\n", i, file, line, pc)
	}

	return buf.String()
}

func (logger *StdFileLogger) LogMode(lv gormlogger.LogLevel) gormlogger.Interface {
	logger.LogLevel = lv
	return logger
}

// info
func (logger *StdFileLogger) Info(ctx context.Context, msg string, data ...interface{}) {
	if logger.LogLevel >= Info {
		logger.printf(logger.infoStr+msg, append([]interface{}{utils.FileWithLineNum()}, data...)...)
	}
}

// warn
func (logger *StdFileLogger) Warn(ctx context.Context, msg string, data ...interface{}) {
	if logger.LogLevel >= Warn {
		logger.printf(logger.infoStr+msg, append([]interface{}{utils.FileWithLineNum()}, data...)...)
	}
}

// error
func (logger *StdFileLogger) Error(ctx context.Context, msg string, data ...interface{}) {
	if logger.LogLevel >= Error {
		//fmt.Println("当前是错误日志:")
		logger.printf("当前是错误日志:\n")
		logger.printf(logger.infoStr+msg, append([]interface{}{utils.FileWithLineNum()}, data...)...)
	}
}
