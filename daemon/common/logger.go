//
// @Author: Jianming Que
// @Description:
// @Version: 1.0.0
// @Date: 2021/2/20 2:38 下午
// @Copyright: MIN-Group；国家重大科技基础设施——未来网络北大实验室；深圳市信息论与未来网络重点实验室
//
package common

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

var (
	_logger       = logrus.New()
	_reportCaller = false
)

//
// 根据字符串返回日志等级枚举值
//
// @Description:
//	返回值 >= 0  => 表示对应的日志等级
//	返回值 = -1  => 表示不输出日志
// @param logLevel
// @return logrus.Level
//
func getLogLevelByString(logLevel string) logrus.Level {
	switch logLevel {
	case "NONE":
		return logrus.FatalLevel
	case "ERROR":
		return logrus.ErrorLevel
	case "WARN":
		return logrus.WarnLevel
	case "INFO":
		return logrus.InfoLevel
	case "DEBUG":
		return logrus.DebugLevel
	case "TRACE":
	case "ALL":
		return logrus.TraceLevel
	}
	return logrus.InfoLevel
}

//
// 根据字符串获取日志的输出格式
//
// @Description:
// @param logFormat
// @return logrus.Formatter
//
func getLogFormatByString(logFormat string) logrus.Formatter {
	switch logFormat {
	case "json":
		return &logrus.JSONFormatter{}
	case "text":
	default:
		return &logrus.TextFormatter{}
	}
	return &logrus.TextFormatter{}
}

//
// 确保某个指定的文件夹存在，存在则直接返回，不存在则创建
//
// @Description:
// @param dirPath		/var/log/mir
// @return error
//
func ensureDirExists(dirPath string) error {
	fi, err := os.Stat(dirPath)

	// 发生错误，且文件夹不存在
	if err != nil && os.IsExist(err) {
		return err
	}

	// 文件或文件夹已存在
	if os.IsExist(err) || err == nil {
		// 文件存在，则判断是否是文件夹，不是文件夹抛出异常
		if fi.IsDir() {
			return nil
		} else {
			return LoggerError{msg: "LogFilePath is a normal file, not a directory!"}
		}
	}

	// 不存在则创建
	return os.Mkdir(dirPath, 0777)
}

//
// 日志模块初始化
//
// @Description:
// @param config		配置文件
//
func InitLogger(config *MIRConfig) {
	_reportCaller = config.LogConfig.ReportCaller
	_logger.SetReportCaller(false)                                         // 日志输出时添加文件名和函数名
	_logger.SetLevel(getLogLevelByString(config.LogConfig.LogLevel))       // 设置日志等级
	_logger.SetFormatter(getLogFormatByString(config.LogConfig.LogFormat)) // 设置日志输出格式

	// 处理指定的文件夹路劲是用户目录开头的情况，例如：~/.mir
	if strings.HasPrefix(config.LogConfig.LogFilePath, "~") {
		homePath, err := Home()
		if err != nil {
			LogFatal("Get current user home path failed!")
		}
		config.LogConfig.LogFilePath = homePath + config.LogConfig.LogFilePath[1:]
	}

	// 设置日志的输出目标
	if config.LogConfig.LogFilePath == "" {
		// 如果为指定路径，则输出到控制台
		_logger.SetOutput(os.Stdout)
	} else {
		// 确保文件夹存在
		if err := ensureDirExists(config.LogConfig.LogFilePath); err != nil {
			LogFatal(err.Error())
		}

		// 根据当前时间戳创建一个本次启动的日志输出
		logFilePath := config.LogConfig.LogFilePath + string(filepath.Separator) + time.Now().Format("2006-1-2.15:04:05") + ".log"
		LogInfo(logFilePath)
		file, err := os.Create(logFilePath)
		if err != nil {
			LogFatal(err.Error())
		}
		_logger.SetOutput(file)
	}
}

func prepareLogger() *logrus.Entry {
	_, file, line, ok := runtime.Caller(2)
	if _reportCaller && ok {
		return _logger.WithFields(logrus.Fields{
			"file": file,
			"line": line,
		})
	} else {
		return _logger.WithFields(logrus.Fields{})
	}
}

// LogLevel: Trace < Debug < Info < Warn < Error < Fatal < Panic

func LogTrace(args ...interface{}) {
	prepareLogger().Trace(args)
}

func LogTraceWithFields(fields logrus.Fields, args ...interface{}) {
	prepareLogger().WithFields(fields).Trace(args)
}

func LogDebug(args ...interface{}) {
	prepareLogger().Debug(args)
}

func LogDebugWithFields(fields logrus.Fields, args ...interface{}) {
	prepareLogger().WithFields(fields).Debug(args)
}

func LogInfo(args ...interface{}) {
	prepareLogger().Info(args)
}

func LogInfoWithFields(fields logrus.Fields, args ...interface{}) {
	prepareLogger().WithFields(fields).Info(args)
}

func LogWarn(args ...interface{}) {
	prepareLogger().Warn(args)
}

func LogWarnWithFields(fields logrus.Fields, args ...interface{}) {
	prepareLogger().WithFields(fields).Warn(args)
}

func LogError(args ...interface{}) {
	prepareLogger().Error(args)
}

func LogErrorWithFields(fields logrus.Fields, args ...interface{}) {
	prepareLogger().WithFields(fields).Error(args)
}

func LogFatal(args ...interface{}) {
	prepareLogger().Fatal(args)
}

func LogFatalWithFields(fields logrus.Fields, args ...interface{}) {
	prepareLogger().WithFields(fields).Fatal(args)
}

func LogPanic(args ...interface{}) {
	prepareLogger().Panic(args)
}

func LogPanicWithFields(fields logrus.Fields, args ...interface{}) {
	prepareLogger().WithFields(fields).Panic(args)
}

/////////////////////////////////////////////////////////////////////////////////////////////////////////
///// 错误处理
/////////////////////////////////////////////////////////////////////////////////////////////////////////

type LoggerError struct {
	msg string
}

func (l LoggerError) Error() string {
	return fmt.Sprintf("LoggerError: %s", l.msg)
}
