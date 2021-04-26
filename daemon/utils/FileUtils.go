// Package utils
// @Author: Jianming Que
// @Description:
// @Version: 1.0.0
// @Date: 2021/4/26 11:27 上午
// @Copyright: MIN-Group；国家重大科技基础设施——未来网络北大实验室；深圳市信息论与未来网络重点实验室
//
package utils

import (
	"minlib/common"
	"strings"
)

func GetRelPath(path string) string {
	if strings.HasPrefix(path, "~") {
		homePath, err := common.Home()
		if err != nil {
			common.LogFatal("Get current user home path failed!")
		}
		return homePath + path[1:]
	} else {
		return path
	}
}
