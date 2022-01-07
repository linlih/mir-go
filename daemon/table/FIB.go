// Copyright [2022] [MIN-Group -- Peking University Shenzhen Graduate School Multi-Identifier Network Development Group]
//
// Licensed under the Apache License, Version 2.0 (the "License"): you may
// not use this file except in compliance with the License. You may obtain
// a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS, WITHOUT
// WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the
// License for the specific language governing permissions and limitations
// under the License.

// Package table
// @Author: yzy
// @Description:
// @Version: 1.0.0
// @Date: 2021/12/21 4:32 PM
// @Copyright: MIN-Group；国家重大科技基础设施——未来网络北大实验室；深圳市信息论与未来网络重点实验室
//

package table

import (
	"github.com/sirupsen/logrus"
	"minlib/common"
	"minlib/component"
	"mir-go/daemon/lf"
)

// FIB
// 储存FIBEntry的前缀树
//
// @Description:
//
type FIB struct {
	lpm     *LpmMatcher //最长前缀匹配器
	version uint64      //版本号
}

// CreateFIB
// 创建并初始化FIB，返回
//
// @Description:
// @return *FIB
//
func CreateFIB() *FIB {
	var f = new(FIB)
	f.lpm = new(LpmMatcher) //初始化
	f.lpm.Create()          //初始化锁
	f.version = 0           // 初始化版本号为0
	return f
}

// Init
// 初始化创建好的FIB表
//
// @Description:
//
func (f *FIB) Init() {
	f.lpm = new(LpmMatcher) //初始化
	f.lpm.Create()          //初始化锁
	f.version = 0
}

// FindLongestPrefixMatch
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

// FindExactMatch
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

// AddOrUpdate
// 通过标识在前缀树中添加或更新FIBEntry 包含NextHop信息
//
// @Description:
// @param *component.Identifier	需要进行查找的标识 logicFaceId  cost 用来创建NextHop的参数
// @return *FIBEntry
//
func (f *FIB) AddOrUpdate(identifier *component.Identifier, logicFace *lf.LogicFace, cost uint64) *FIBEntry {
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
		entry.SetIdentifier(identifier)
		entry.NextHopList[logicFace.LogicFaceId] = &NextHop{LogicFace: logicFace, Cost: cost}
		return entry
	})
	f.version++
	return val.(*FIBEntry)
}

// EraseByIdentifier
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
	f.version++
	return f.lpm.Delete(PrefixList)
}

// EraseByFIBEntry
// 通过FIBEntry在前缀树中删除FIBEntry
//
// @Description:
// @param *FIBEntry	需要删除的FIBEntry
// @return error
//
func (f *FIB) EraseByFIBEntry(fibEntry *FIBEntry) error {
	var PrefixList []string
	for _, v := range fibEntry.GetIdentifier().GetComponents() {
		PrefixList = append(PrefixList, v.ToString())
	}
	f.version++
	return f.lpm.Delete(PrefixList)
}

// RemoveNextHopByFace
// 删除FIB表中，所有以logicFaceId为下一跳的表项 返回删除的表项数
//
// @Description:
// @param logicFaceId
// @return uint64
//
func (f *FIB) RemoveNextHopByFace(logicFace *lf.LogicFace) uint64 {
	f.version++
	return f.lpm.TraverseFunc(func(val interface{}) uint64 {
		if v, ok := val.(*FIBEntry); ok {
			if v.HasNextHop(logicFace) {
				v.RemoveNextHop(logicFace)
				return 1
			}
		} else {
			common.LogErrorWithFields(logrus.Fields{
				"value": val,
			}, "FIBEntry transform fail")
		}
		return 0
	})
}

// Size
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
			common.LogErrorWithFields(logrus.Fields{
				"value": val,
			}, "FIBEntry transform fail")
		}
		return 0
	})
}

// GetDepth
// 返回前缀树当前的深度
//
// @Description: 返回前缀树当前的深度
// @return int
//
func (f *FIB) GetDepth() int {
	// 根节点不存储数据
	return f.lpm.GetDepth() - 1
}

// GetAllEntry
// 返回FIB表中所有的表项
//
// @Description:返回FIB表中所有的表项
// @return []*FIBEntry
//
func (f *FIB) GetAllEntry() []*FIBEntry {
	var fibEntries []*FIBEntry
	f.lpm.TraverseFunc(func(val interface{}) uint64 {
		if fibEntry, ok := val.(*FIBEntry); ok {
			fibEntries = append(fibEntries, fibEntry)
			return 1
		} else {
			common.LogErrorWithFields(logrus.Fields{
				"value": val,
			}, "FIBEntry transform fail")
		}
		return 0
	})
	return fibEntries
}

// GetVersion
// 返回FIB表当前的版本号
//
// @Description:返回FIB表当前的版本号
// @return uint64
//
func (f *FIB) GetVersion() uint64 {
	return f.version
}
