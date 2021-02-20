//
// @Author: Jianming Que
// @Description:
// @Version: 1.0.0
// @Date: 2021/2/20 2:38 下午
// @Copyright: MIN-Group；国家重大科技基础设施——未来网络北大实验室；深圳市信息论与未来网络重点实验室
//
package common

import "fmt"

//
// 打印日志
// @Description:
// @param msg
//
func LogFatal(msg string) {
	fmt.Printf("FATAL: %s", msg)
}
