/**
 * @Author: yzy
 * @Description:
 * @Version: 1.0.0
 * @Date: 2021/3/4 上午12:48
 *@Copyright: MIN-Group；国家重大科技基础设施——未来网络北大实验室；深圳市信息论与未来网络重点实验室
 */

package table

import (
	"minlib/component"
	"sort"
	"sync"
)

//下一跳
type NextHop struct {
	LogicFaceId uint64 //逻辑接口号
	Cost        uint64 //路由开销
}

/*
FIBEntry类用于存储FIB每条表项的具体内容
类的属性成员至少包括
Identifier、表项有效期和下一跳列表
而下一跳列表中的每一项应包含下一跳逻辑接口号（uint64）和路由开销
*/
type FIBEntry struct {
	Identifier *component.Identifier //标识对象指针
	//Ttl         time.Duration         //表项有效期
	NextHopList map[uint64]NextHop //下一跳列表 用map实现是为了查找和删除方便
	RWlock      *sync.RWMutex      //读写锁
}

func CreateFIBEntry() *FIBEntry {
	var f = &FIBEntry{}
	f.RWlock = new(sync.RWMutex)
	f.NextHopList = make(map[uint64]NextHop)
	return f
}

//获得标识
func (f *FIBEntry) GetIdentifier() *component.Identifier { return f.Identifier }

//获得下一跳列表 列表应该按cost从小到大排序
func (f *FIBEntry) GetNextHops() []NextHop {
	NextHopList := make([]NextHop, 0)
	//if f.NextHopList == nil {
	//	f.NextHopList = make(map[uint64]NextHop)
	//	return NextHopList
	//}
	f.RWlock.RLock()
	for _, nextHop := range f.NextHopList {
		NextHopList = append(NextHopList, nextHop)
	}
	f.RWlock.RUnlock()
	// 内置函数 按照cost从小到大排序
	sort.Slice(NextHopList, func(i, j int) bool {
		return NextHopList[i].Cost < NextHopList[j].Cost
	})
	return NextHopList
}

//判断有没有下一跳的信息 true表示有数据 false表示没有数据
func (f *FIBEntry) HasNextHops() bool {
	//if f.NextHopList == nil {
	//	f.NextHopList = make(map[uint64]NextHop)
	//}
	return len(f.NextHopList) != 0
}

// 判断logicFaceId是否在下一跳列表中
func (f *FIBEntry) HasNextHop(logicFaceId uint64) bool {
	f.RWlock.RLock()
	//for _, nextHop := range f.NextHopList {
	//	if nextHop.LogicFaceId == logicFaceId {
	//		return true
	//	}
	//}
	_, ok := f.NextHopList[logicFaceId]
	f.RWlock.RUnlock()
	return ok
}

// 添加或更新一个下一跳信息
func (f *FIBEntry) AddOrUpdateNextHop(logicFaceId uint64, cost uint64) {
	//if f.NextHopList == nil {
	//	f.NextHopList = make(map[uint64]NextHop)
	//}
	f.RWlock.Lock()
	delete(f.NextHopList, logicFaceId)
	f.NextHopList[logicFaceId] = NextHop{LogicFaceId: logicFaceId, Cost: cost}
	f.RWlock.Unlock()
}

// 删除一个下一跳信息
func (f *FIBEntry) RemoveNextHop(logicFaceId uint64) {
	//if f.NextHopList == nil {
	//	f.NextHopList = make(map[uint64]NextHop)
	//	return
	//}
	f.RWlock.Lock()
	delete(f.NextHopList, logicFaceId)
	f.RWlock.Unlock()
}
