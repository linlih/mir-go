// Package lf
// @Author: weiguohua
// @Description:
// @Version: 1.0.0
// @Date: 2021/3/14 下午10:05
// @Copyright: MIN-Group；国家重大科技基础设施——未来网络北大实验室；深圳市信息论与未来网络重点实验室
//
package lf

import (
	"minlib/utils"
)

// LogicFaceTable
// @Description:  用于保存LogicFaceId和真正Face对象的映射关系。
//
type LogicFaceTable struct {
	mLogicFaceTable LogicFaceMap
	mSize           utils.ThreadFreeUint64 // LogicFace 表的大小
	lastId          utils.ThreadFreeUint64 // 下一个分配的 LogicFace Id
	version         utils.ThreadFreeUint64 // 版本
	OnEvicted       func(uint64)
}

// Init 初始化 LogicFace Table
//
// @Description:
// @receiver l
//
func (l *LogicFaceTable) Init() {
	l.lastId.SetValue(0)
	l.mSize.SetValue(0)
	l.version.SetValue(0)
}

// GetVersion 获取版本号
//
// @Description:
// @receiver l
// @return uint64
//
func (l *LogicFaceTable) GetVersion() uint64 {
	return l.version.GetValue()
}

// Size
// @Description: 获得当前表大小。
// @receiver logicFaceTable
// @return uint64  表中LogicFace个数
//
func (l *LogicFaceTable) Size() uint64 {
	return l.mSize.GetValue()
}

// AddLogicFace
// @Description: 往LogicFaceTable添加一个LogicFace
// @receiver logicFaceTable
// @param logicFacePtr LogicFace对象指针
// @return uint64  返回分配的LogicFaceId
//
func (l *LogicFaceTable) AddLogicFace(logicFacePtr *LogicFace) uint64 {
	logicFacePtr.LogicFaceId = l.lastId.GetAndPlus(1)
	l.mLogicFaceTable.StoreLogicFace(logicFacePtr.LogicFaceId, logicFacePtr)
	l.mSize.AddAndGet(1)
	l.version.AddAndGet(1)
	logicFacePtr.SetOnShutdownCallback(func(logicFaceId uint64) {
		l.OnEvicted(logicFaceId)
	})
	return logicFacePtr.LogicFaceId
}

// GetLogicFacePtrById
// @Description: 通过LogicFaceId来获得LogicFace对象指针。
// @receiver logicFaceTable
// @param logicFaceId 	logicFace号
// @return *LogicFace	LogicFace对象指针
//
func (l *LogicFaceTable) GetLogicFacePtrById(logicFaceId uint64) *LogicFace {
	return l.mLogicFaceTable.LoadLogicFace(logicFaceId)
}

// RemoveByLogicFaceId
// @Description:  通过LogicFaceId来删除某个表项。
// @receiver logicFaceTable
// @param logicFaceId logicFace号
//
func (l *LogicFaceTable) RemoveByLogicFaceId(logicFaceId uint64) {
	l.mLogicFaceTable.Delete(logicFaceId)
	l.mSize.SubtractAndGet(1)
	l.version.AddAndGet(1)
}

// Range 遍历 LogicFace Table
//
// @Description:
// @receiver l
// @param f
//
func (l *LogicFaceTable) Range(f func(key uint64, value *LogicFace) bool) {
	l.mLogicFaceTable.Range(func(key, value interface{}) bool {
		return f(key.(uint64), value.(*LogicFace))
	})
}

// GetAllFaceList
// @Description:  获取所有face表项
// @return []*LogicFace 逻辑face列表
//
func (l *LogicFaceTable) GetAllFaceList() []*LogicFace {
	var faceList []*LogicFace
	l.mLogicFaceTable.Range(func(key, value interface{}) bool {
		faceList = append(faceList, value.(*LogicFace))
		return true
	})
	return faceList
}
