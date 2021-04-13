// Package common
// @Author: Jianming Que
// @Description:
// @Version: 1.0.0
// @Date: 2021/2/20 2:38 下午
// @Copyright: MIN-Group；国家重大科技基础设施——未来网络北大实验室；深圳市信息论与未来网络重点实验室
//
package common

import (
	"minlib/common"
)

// InitLogger 日志模块初始化
//
// @Description:
// @param config		配置文件
//
func InitLogger(config *MIRConfig) {
	common.InitLogger(&common.LoggerParameters{
		ReportCaller: config.LogConfig.ReportCaller,
		LogLevel:     config.LogConfig.LogLevel,
		LogFormat:    config.LogConfig.LogFormat,
		LogFilePath:  config.LogConfig.LogFilePath,
	})
}
