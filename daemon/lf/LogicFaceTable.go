// Package lf
// @Author: weiguohua
// @Description:
// @Version: 1.0.0
// @Date: 2021/3/14 下午10:05
// @Copyright: MIN-Group；国家重大科技基础设施——未来网络北大实验室；深圳市信息论与未来网络重点实验室
//
package lf

import "sync"

// LogicFaceTable
// @Description:  用于保存LogicFaceId和真正Face对象的映射关系。
//
type LogicFaceTable struct {
	mLogicFaceTable map[uint64]*LogicFace
	mSize           uint64
	tableLock       sync.Mutex
	lastId          uint64
	version         uint64 // 版本
}

// GetVersion 获取版本号
//
// @Description:
// @receiver l
// @return uint64
//
func (l *LogicFaceTable) GetVersion() uint64 {
	return l.version
}

func (l *LogicFaceTable) Init() {
	l.lastId = 0
	l.mLogicFaceTable = make(map[uint64]*LogicFace)
	l.mSize = 0
	l.version = 0
}

// AddLogicFace
// @Description: 往LogicFaceTable添加一个LogicFace
// @receiver logicFaceTable
// @param logicFacePtr LogicFace对象指针
// @return uint64  返回分配的LogicFaceId
//
func (l *LogicFaceTable) AddLogicFace(logicFacePtr *LogicFace) uint64 {
	lfid := l.lastId
	l.tableLock.Lock()
	l.mLogicFaceTable[l.lastId] = logicFacePtr
	logicFacePtr.LogicFaceId = l.lastId
	l.mSize++
	l.lastId++
	l.version++
	l.tableLock.Unlock()
	return lfid
}

// GetLogicFacePtrById
// @Description: 通过LogicFaceId来获得LogicFace对象指针。
// @receiver logicFaceTable
// @param logicFaceId 	logicFace号
// @return *LogicFace	LogicFace对象指针
//
func (l *LogicFaceTable) GetLogicFacePtrById(logicFaceId uint64) *LogicFace {
	var logicFacePtr *LogicFace = nil
	l.tableLock.Lock()
	logicFacePtr = l.mLogicFaceTable[logicFaceId]
	l.tableLock.Unlock()
	return logicFacePtr
}

// Size
// @Description: 获得当前表大小。
// @receiver logicFaceTable
// @return uint64  表中LogicFace个数
//
func (l *LogicFaceTable) Size() uint64 {
	return uint64(len(l.mLogicFaceTable))
}

// RemoveByLogicFaceId
// @Description:  通过LogicFaceId来删除某个表项。
// @receiver logicFaceTable
// @param logicFaceId logicFace号
//
func (l *LogicFaceTable) RemoveByLogicFaceId(logicFaceId uint64) {
	l.tableLock.Lock()
	delete(l.mLogicFaceTable, logicFaceId)
	l.mSize--
	l.version--
	l.tableLock.Unlock()

}

// GetAllFaceList
// @Description:  获取所有face表项
// @return []*LogicFace 逻辑face列表
//
func (l *LogicFaceTable) GetAllFaceList() []*LogicFace {
	var faceList []*LogicFace
	for _, v := range l.mLogicFaceTable {
		faceList = append(faceList, v)
	}
	return faceList
}
