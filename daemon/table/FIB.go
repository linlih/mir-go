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

func (f *FIB) FindLongestPrefixMatch(identifier *component.Identifier) *FIBEntry {
	var PrefixList []string
	for _, v := range identifier.GetComponents() {
		PrefixList = append(PrefixList, v.ToString())
	}
	if table, ok := f.lpm.FindLongestPrefixMatch(PrefixList); ok {
		return table.(*FIBEntry)
	}
	// 匹配失败返回空
	return nil
}

func (f *FIB) FindExactMatch(identifier *component.Identifier) *FIBEntry {
	var PrefixList []string
	for _, v := range identifier.GetComponents() {
		PrefixList = append(PrefixList, v.ToString())
	}
	if table, ok := f.lpm.FindExactMatch(PrefixList); ok {
		return table.(*FIBEntry)
	}
	// 匹配失败返回空
	return nil
}

func (f *FIB) AddOrUpdate(identifier *component.Identifier, logicFaceId uint64, cost uint64) *FIBEntry {
	var PrefixList []string
	for _, v := range identifier.GetComponents() {
		PrefixList = append(PrefixList, v.ToString())
	}
	//两种模式 模式一 val == nil func not nil 用处理函数设置val 模式二 val not nil func == nil 直接设置val
	val := f.lpm.AddOrUpdate(PrefixList, nil, func(val interface{}) interface{} {
		// not ok 那么val == nil 存入标识
		if _, ok := (val).(*FIBEntry); !ok {
			fibEntry := CreateFIBEntry()
			val = fibEntry
		}
		entry := (val).(*FIBEntry)
		entry.Identifier = identifier
		entry.NextHopList[logicFaceId] = NextHop{LogicFaceId: logicFaceId, Cost: cost}
		return entry
	})
	return val.(*FIBEntry)
}

func (f *FIB) EraseByIdentifier(identifier *component.Identifier) error {
	var PrefixList []string
	for _, v := range identifier.GetComponents() {
		PrefixList = append(PrefixList, v.ToString())
	}
	return f.lpm.Delete(PrefixList)
}

func (f *FIB) EraseByFIBEntry(fibEntry *FIBEntry) error {
	var PrefixList []string
	for _, v := range fibEntry.Identifier.GetComponents() {
		PrefixList = append(PrefixList, v.ToString())
	}
	return f.lpm.Delete(PrefixList)
}

func (f *FIB) RemoveNextHopByFace(logicFaceId uint64) uint64 {
	return f.lpm.TraverseFunc(func(val interface{}) uint64 {
		if v, ok := val.(*FIBEntry); ok {
			if v.HasNextHop(logicFaceId) {
				v.RemoveNextHop(logicFaceId)
				return 1
			}
		} else {
			fmt.Println("FIBEntry transform fail")
		}
		return 0
	})
}

func (f *FIB) Size() uint64 {
	return f.lpm.TraverseFunc(func(val interface{}) uint64 {
		if _, ok := val.(*FIBEntry); ok {
			return 1
		} else {
			fmt.Println("FIBEntry transform fail")
		}
		return 0
	})
}
