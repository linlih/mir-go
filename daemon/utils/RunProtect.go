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

// Package utils
// @Author: Jianming Que
// @Description:
// @Version: 1.0.0
// @Date: 2021/12/21 4:32 PM
// @Copyright: MIN-Group；国家重大科技基础设施——未来网络北大实验室；深圳市信息论与未来网络重点实验室
//
package utils

import (
	"minlib/common"
)

func ProtectRun(work func(), onError func(err interface{})) {
	//defer func() {
	//	if err := recover(); err != nil {
	//		onError(err)
	//	}
	//}()
	work()
}

// GoroutineNoPanic 在一个单独的协程里面运行一个方法，并且不会panic异常
//
// @Description:
// @param work
//
func GoroutineNoPanic(work func()) {
	go ProtectRun(work, func(err interface{}) {
		common.LogError(err)
	})
}
