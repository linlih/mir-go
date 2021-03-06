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
	"fmt"
	"github.com/sirupsen/logrus"
	common2 "minlib/common"
	"minlib/packet"
	"mir-go/daemon/lf"
)

// PIT
// PIT表结构体
//
// @Description:PIT表结构体,用前缀树存储表项
//
type PIT struct {
	lpm *LpmMatcher //最长前缀匹配器
}

// CreatePIT
// 初始化PIT表 并返回
//
// @Description:
// @return *PIT
//
func CreatePIT() *PIT {
	var p = &PIT{}
	p.lpm = &LpmMatcher{} //初始化
	p.lpm.Create()        //初始化锁
	return p
}

// Init
// 初始化创建好的PIT表
//
// @Description:
//
func (p *PIT) Init() {
	p.lpm = new(LpmMatcher) //初始化
	p.lpm.Create()          //初始化锁
}

// Size
// 返回PIT表中含有数据的表项数
//
// @Description:
// @return uint64
//
func (p *PIT) Size() uint64 {
	return p.lpm.TraverseFunc(func(val interface{}) uint64 {
		if _, ok := val.(*PITEntry); ok {
			return 1
		} else {
			common2.LogErrorWithFields(logrus.Fields{
				"value": val,
			}, "PITEntry transform fail")
		}
		return 0
	})
}

// Find
// 通过兴趣包在前缀树中精准匹配查找对应的PITEntry
//
// @Description:
// @param *packet.Interest	需要进行查找的兴趣包
// @return *PITEntry error
//
func (p *PIT) Find(interest *packet.Interest) (*PITEntry, error) {
	var PrefixList []string
	for _, v := range interest.GetName().GetComponents() {
		PrefixList = append(PrefixList, v.ToString())
	}
	if v, ok := p.lpm.FindExactMatch(PrefixList); ok {
		if pitEntry, ok := v.(*PITEntry); ok {
			return pitEntry, nil
		}
	}
	return nil, createPITErrorByType(PITEntryNotExistedError)
}

// Insert
// 在PIT表中插入PITEntry
//
// @Description:
// @param *packet.Interest 兴趣包指针
// @return *PITEntry
//
func (p *PIT) Insert(interest *packet.Interest) *PITEntry {
	var PrefixList []string
	for _, v := range interest.GetName().GetComponents() {
		PrefixList = append(PrefixList, v.ToString())
	}
	val := p.lpm.AddOrUpdate(PrefixList, nil, func(val interface{}) interface{} {
		// not ok 那么val == nil 存入标识
		if _, ok := (val).(*PITEntry); !ok {
			// 存入的表项 不是 *PITEntry类型 或者 为nil
			pitEntry := CreatePITEntry()
			val = pitEntry
		}
		entry := (val).(*PITEntry)
		entry.Identifier = interest.GetName()
		return entry
	})
	return val.(*PITEntry)
}

// FindDataMatches
// 根据数据包在PIT表中获取PITEntry 匹配过程应该在PIT表中删除匹配过的表项
//
// @Description:
// @param *packet.Data	数据包指针
// @return *PITEntry
//
func (p *PIT) FindDataMatches(data *packet.Data) *PITEntry {
	var PrefixList []string
	for _, v := range data.GetName().GetComponents() {
		PrefixList = append(PrefixList, v.ToString())
	}
	if v, ok := p.lpm.FindExactMatch(PrefixList); ok {
		p.lpm.Delete(PrefixList)
		return v.(*PITEntry)
	}
	return nil
}

// EraseByPITEntry
// 根据PITEntry删除PIT表中的PITEntry
//
// @Description:
// @param *PITEntry
// @return error
//
func (p *PIT) EraseByPITEntry(pitEntry *PITEntry) error {
	var PrefixList []string
	for _, v := range pitEntry.Identifier.GetComponents() {
		PrefixList = append(PrefixList, v.ToString())
	}
	return p.lpm.Delete(PrefixList)
}

// EraseByLogicFace
// 删除所有以logicFace为流入接口号或流出接口号的表项,返回删除的数量
//
// @Description:
// @param logicFace
// @return uint64
//
func (p *PIT) EraseByLogicFace(logicFace *lf.LogicFace) uint64 {
	return p.lpm.TraverseFunc(func(val interface{}) uint64 {
		if pitEntry, ok := val.(*PITEntry); ok {
			var ok1, ok2 bool
			if _, ok1 = pitEntry.InRecordList[logicFace.LogicFaceId]; ok1 {
				delete(pitEntry.InRecordList, logicFace.LogicFaceId)
			}
			if _, ok2 = pitEntry.OutRecordList[logicFace.LogicFaceId]; ok2 {
				delete(pitEntry.OutRecordList, logicFace.LogicFaceId)
			}
			if ok1 || ok2 {
				return 1
			}
			return 0
		}
		common2.LogErrorWithFields(logrus.Fields{
			"value": val,
		}, "PITEntry transform fail")
		return 0
	})
}

/////////////////////////////////////////////////////////////////////////////////////////////////////////
///// 错误处理
/////////////////////////////////////////////////////////////////////////////////////////////////////////

const (
	PITEntryNotExistedError = iota
)

type PITError struct {
	msg string
}

func (p PITError) Error() string {
	return fmt.Sprintf("NodeError: %s", p.msg)
}

func createPITErrorByType(errorType int) (err PITEntryError) {
	switch errorType {
	case PITEntryNotExistedError:
		err.msg = "PITEntry not found by interest"
	default:
		err.msg = "Unknown error"
	}
	return
}
