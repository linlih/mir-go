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
	"mir-go/daemon/lf"
	"sort"
	"sync"
)

// NextHop
// 下一跳结构体
//
// @Description:下一跳结构体 用于存储下一跳信息
//
type NextHop struct {
	LogicFace *lf.LogicFace //逻辑接口号
	Cost      uint64        //路由开销
}

// FIBEntry
// FIBEntry结构体 对应FIB表项
//
// @Description:
//	1.FIBEntry类用于存储FIB每条表项的具体内容
//	2.类的属性成员至少包括Identifier、表项有效期和下一跳列表
//	3.下一跳列表中的每一项应包含下一跳逻辑接口号（uint64）和路由开销
//
type FIBEntry struct {
	identifier  *component.Identifier //标识对象指针
	NextHopList map[uint64]*NextHop   //下一跳列表 用map实现是为了查找和删除方便
	readOnly    bool                  // 设置该fibEntry是否为只读
	RWlock      *sync.RWMutex         //读写锁
}

// CreateFIBEntry
// 初始化FIBEntry并返回
//
// @Description:
// @return *FIBEntry
//
func CreateFIBEntry() *FIBEntry {
	return &FIBEntry{
		RWlock:      new(sync.RWMutex),
		NextHopList: make(map[uint64]*NextHop),
		readOnly:    false,
	}
}

// GetIdentifier
// 返回FIBEntry的标识
//
// @Description:
// @return *component.Identifier
//
func (f *FIBEntry) GetIdentifier() *component.Identifier { return f.identifier }

func (f *FIBEntry) SetIdentifier(identifier *component.Identifier) { f.identifier = identifier }

// GetNextHops
// 返回FIBEntry中的下一跳列表 列表应该按cost从小到大排序
//
// @Description:
// @return []NextHop
//
func (f *FIBEntry) GetNextHops() []*NextHop {
	NextHopList := make([]*NextHop, 0)
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

// HasNextHops
// 判断有没有下一跳的信息 true表示有数据 false表示没有数据
//
// @Description:
// @return bool
//
func (f *FIBEntry) HasNextHops() bool {
	return len(f.NextHopList) != 0
}

// HasNextHop
// 判断logicFaceId是否在下一跳列表中
//
// @Description:
// @return bool
//
func (f *FIBEntry) HasNextHop(logicFace *lf.LogicFace) bool {
	f.RWlock.RLock()
	_, ok := f.NextHopList[logicFace.LogicFaceId]
	f.RWlock.RUnlock()
	return ok
}

// AddOrUpdateNextHop
// 添加或更新下一跳信息
//
// @Description:
// @param logicFaceId,cost 下一跳信息
//
func (f *FIBEntry) AddOrUpdateNextHop(logicFace *lf.LogicFace, cost uint64) {
	f.RWlock.Lock()
	delete(f.NextHopList, logicFace.LogicFaceId)
	f.NextHopList[logicFace.LogicFaceId] = &NextHop{LogicFace: logicFace, Cost: cost}
	f.RWlock.Unlock()
}

// RemoveNextHop
// 删除下一跳信息
//
// @Description:
// @param logicFaceId 下一跳信息的logicFaceId
//
func (f *FIBEntry) RemoveNextHop(logicFace *lf.LogicFace) {
	f.RWlock.Lock()
	delete(f.NextHopList, logicFace.LogicFaceId)
	f.RWlock.Unlock()
}

// SetReadOnly
// 设置只读
//
// @Description: 设置只读
//
func (f *FIBEntry) SetReadOnly() {
	f.readOnly = true
}

// IsChanged
// FIBEntry是否可变
//
// @Description: FIBEntry是否可变
// @Return: bool
//
func (f *FIBEntry) IsChanged() bool {
	return !f.readOnly
}
