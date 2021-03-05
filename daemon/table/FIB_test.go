/**
 * @Author: wzx
 * @Description:
 * @Version: 1.0.0
 * @Date: 2021/3/4 上午2:05
 *@Copyright: MIN-Group；国家重大科技基础设施——未来网络北大实验室；深圳市信息论与未来网络重点实验室
 */

package table

import (
	"fmt"
	"testing"
)
import "minlib/component"

func TestFib(t *testing.T){
	var fib FIB
	fib.Create()
	fib.Insert(component.Identifier{},1, 1)

	fmt.Println(fib.FindExactMatch(&component.Identifier{}))
}