//
// @Author: weiguohua
// @Description:
// @Version: 1.0.0
// @Date: 2021/3/23 下午3:08
// @Copyright: MIN-Group；国家重大科技基础设施——未来网络北大实验室；深圳市信息论与未来网络重点实验室
//
package common

import "time"

//
// 获取当前的时间戳（单位为 ms）
//
// @Description:
// @return uint64
//
func GetCurrentTime() uint64 {
	return uint64(time.Now().UnixNano() / 1e6)
}
