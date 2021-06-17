package common

import (
	"fmt"
	"testing"
)

//
// @Author: Wei Guohua
// @Description:
// @Version: 1.0.0
// @Date: 2021/6/17 上午10:39
// @Copyright: MIN-Group；国家重大科技基础设施——未来网络北大实验室；深圳市信息论与未来网络重点实验室
//

func Test_ParseDefaultConfig(t *testing.T) {
	config, err := ParseDefaultConfig("../../defaultRoute.xml")
	fmt.Println(config)
	fmt.Println(err)
}
