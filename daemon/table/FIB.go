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

//
// 储存FIBEntry的前缀树
//
// @Description:
//
type FIB struct {
	lpm *LpmMatcher //最长前缀匹配器
}

//
// 创建并初始化FIB，返回
//
// @Description:
// @return *FIB
//
func CreateFIB() *FIB {
	var f = &FIB{}
	f.lpm = &LpmMatcher{} //初始化
	f.lpm.Create()        //初始化锁
	return f
}

//
// 通过标识在前缀树中最长前缀匹配查找对应的FIBEntry 最长前缀匹配的意思是有尽量多个Component可以匹配到结果
//
// @Description:
// @param *component.Identifier	需要进行查找的标识
// @return *FIBEntry
//
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

//
// 通过标识在前缀树中准确匹配查找对应的FIBEntry
//
// @Description:
// @param *component.Identifier	需要进行查找的标识
// @return *FIBEntry
//
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

//
// 通过标识在前缀树中添加或更新FIBEntry 包含NextHop信息
//
// @Description:
// @param *component.Identifier	需要进行查找的标识 logicFaceId  cost 用来创建NextHop的参数
// @return *FIBEntry
//
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

//
// 通过标识在前缀树中删除FIBEntry
//
// @Description:
// @param *component.Identifier	需要删除的标识
// @return error
//
func (f *FIB) EraseByIdentifier(identifier *component.Identifier) error {
	var PrefixList []string
	for _, v := range identifier.GetComponents() {
		PrefixList = append(PrefixList, v.ToString())
	}
	return f.lpm.Delete(PrefixList)
}

//
// 通过FIBEntry在前缀树中删除FIBEntry
//
// @Description:
// @param *FIBEntry	需要删除的FIBEntry
// @return error
//
func (f *FIB) EraseByFIBEntry(fibEntry *FIBEntry) error {
	var PrefixList []string
	for _, v := range fibEntry.Identifier.GetComponents() {
		PrefixList = append(PrefixList, v.ToString())
	}
	return f.lpm.Delete(PrefixList)
}

//
// 删除FIB表中，所有以logicFaceId为下一跳的表项 返回删除的表项数
//
// @Description:
// @param logicFaceId
// @return uint64
//
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

//
// 返回前缀树里存有FIBEntry的节点数
//
// @Description:
// @return uint64
//
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
