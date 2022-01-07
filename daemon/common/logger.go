// Copyright [2022] [MIN-Group -- Peking University Shenzhen Graduate School Multi-Identifier Network Development Group]
//
// Licensed under the Apache License, Version 2.0 (the "License"): you may
// not use this file except in compliance with the License. You may obtain
// a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS, WITHOUT
// WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the
// License for the specific language governing permissions and limitations
// under the License.

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
