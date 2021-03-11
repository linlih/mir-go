/**
 * @Author: yzy
 * @Description:
 * @Version: 1.0.0
 * @Date: 2021/3/4 上午1:37
 *@Copyright: MIN-Group；国家重大科技基础设施——未来网络北大实验室；深圳市信息论与未来网络重点实验室
 */

package table

import (
	"fmt"
	"minlib/component"
)

type FIB struct {
	lpm *LpmMatcher //最长前缀匹配器
}

func CreateFIB() *FIB {
	var f = &FIB{}
	f.lpm = &LpmMatcher{} //初始化
	f.lpm.Create()        //初始化锁
	return f
}


