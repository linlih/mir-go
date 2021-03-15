//
// @Author: Jianming Que
// @Description:
// @Version: 1.0.0
// @Date: 2021/2/20 2:38 下午
// @Copyright: MIN-Group；国家重大科技基础设施——未来网络北大实验室；深圳市信息论与未来网络重点实验室
//
package common

import (
	"github.com/sirupsen/logrus"
	"os"
)

var (
	_logger = logrus.New()
)

//
// 日志模块初始化
// TODO: 在这里应该读取配置文件，然后初始化日志模块
//
// @Description:
//
func InitLogger() {
	_logger.SetReportCaller(true)                 // 日志输出时添加文件名名和函数名
	_logger.SetLevel(logrus.TraceLevel)           // 设置日志等级
	_logger.SetFormatter(&logrus.TextFormatter{}) // 设置日志输出格式
	_logger.SetOutput(os.Stdout)                  // 设置日志的输出目标
}

// LogLevel: Trace < Debug < Info < Warn < Error < Fatal < Panic

func LogTrace(args ...interface{}) {
	_logger.Trace(args)
}

func LogTraceWithFields(fields logrus.Fields, args ...interface{}) {
	_logger.WithFields(fields).Trace(args)
}

func LogDebug(args ...interface{}) {
	_logger.Debug(args)
}

func LogDebugWithFields(fields logrus.Fields, args ...interface{}) {
	_logger.WithFields(fields).Debug(args)
}

func LogInfo(args ...interface{}) {
	_logger.Info(args)
}

func LogInfoWithFields(fields logrus.Fields, args ...interface{}) {
	_logger.WithFields(fields).Info(args)
}

func LogWarn(args ...interface{}) {
	_logger.Warn(args)
}

func LogWarnWithFields(fields logrus.Fields, args ...interface{}) {
	_logger.WithFields(fields).Warn(args)
}

func LogError(args ...interface{}) {
	_logger.Error(args)
}

func LogErrorWithFields(fields logrus.Fields, args ...interface{}) {
	_logger.WithFields(fields).Error(args)
}

func LogFatal(args ...interface{}) {
	_logger.Fatal(args)
}

func LogFatalWithFields(fields logrus.Fields, args ...interface{}) {
	_logger.WithFields(fields).Fatal(args)
}

func LogPanic(args ...interface{}) {
	_logger.Panic(args)
}

func LogPanicWithFields(fields logrus.Fields, args ...interface{}) {
	_logger.WithFields(fields).Panic(args)
}
