/**
 * @Author: wzx
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
	lpm    *LpmMatcher
}

func(f *FIB) Create(){
	f.lpm = &LpmMatcher{}
	f.lpm.Create()
}

func(f *FIB)Insert(identifier component.Identifier, logicFaceId uint64, cost uint64 ) {
	s := make([]string, 0)
	s = append(s, "test")
	f.lpm.AddOrUpdate(s, nil, func(val interface{}) interface{} {
		if _, ok :=(val).(*FIBEntry); !ok{
			fmt.Println("1111")
			val = &FIBEntry{Identifier: &identifier}
		}
		entry := (val).(*FIBEntry)
		if entry.NextHopList == nil{
			entry.NextHopList = make(map[uint64]NextHop)
		}
		entry.NextHopList[logicFaceId] = NextHop{LogicFaceId: logicFaceId, Cost: cost}

		return entry
	})
}

func(f *FIB)FindExactMatch(identifier *component.Identifier) *FIBEntry{
	s := make([]string, 0)
	s = append(s, "test")
	entry, _ := f.lpm.FindExactMatch(s)
	return  entry.(*FIBEntry)

}

