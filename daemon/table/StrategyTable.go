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
	common2 "minlib/common"
	"minlib/component"
)

type StrategyTable struct {
	lpm *LpmMatcher //最长前缀匹配器
}

func CreateStrategyTable() *StrategyTable {
	var s = &StrategyTable{}
	s.lpm = &LpmMatcher{} //初始化
	s.lpm.Create()        //初始化锁
	return s
}

func (s *StrategyTable) Init() {
	s.lpm = &LpmMatcher{} //初始化
	s.lpm.Create()        //初始化锁
}

// Size 获得StrategyTable的大小
func (s *StrategyTable) Size() uint64 {
	return s.lpm.TraverseFunc(func(val interface{}) uint64 {
		if _, ok := val.(*StrategyTableEntry); ok {
			return 1
		} else {
			common2.LogErrorWithFields(logrus.Fields{
				"value": val,
			}, "StrategyTableEntry transform fail")
		}
		return 0
	})
}

// SetDefaultStrategy 为所有的前缀设置一个默认的策略
func (s *StrategyTable) SetDefaultStrategy(strategyName string) {
	s.lpm.TraverseFunc(func(val interface{}) uint64 {
		if strategyTableEntry, ok := val.(*StrategyTableEntry); ok {
			strategyTableEntry.StrategyName = strategyName
			return 1
		} else {
			common2.LogErrorWithFields(logrus.Fields{
				"value": val,
			}, "StrategyTableEntry transform fail")
		}
		return 0
	})
}

// Insert 往策略表中插入一个策略
func (s *StrategyTable) Insert(identifier *component.Identifier, strategyName string, istrategy IStrategy) *StrategyTableEntry {
	var PrefixList []string
	for _, v := range identifier.GetComponents() {
		PrefixList = append(PrefixList, v.ToString())
	}
	val := s.lpm.AddOrUpdate(PrefixList, nil, func(val interface{}) interface{} {
		// not ok 那么val == nil 存入标识
		if _, ok := (val).(*StrategyTableEntry); !ok {
			// 存入的表项 不是 *StrategyTableEntry类型 或者 为nil
			strategyTableEntry := CreateStrategyTableEntry()
			val = strategyTableEntry
		}
		entry := (val).(*StrategyTableEntry)
		entry.StrategyName = strategyName
		entry.IStrategy = istrategy
		return entry
	})
	return val.(*StrategyTableEntry)

}

// Erase 通过前缀删除策略表中策略
func (s *StrategyTable) Erase(identifier *component.Identifier) error {
	var PrefixList []string
	for _, v := range identifier.GetComponents() {
		PrefixList = append(PrefixList, v.ToString())
	}
	return s.lpm.Delete(PrefixList)
}

// FindEffectiveStrategyEntry 查询和一个指定的名称前缀匹配的策略条目 最长前缀匹配
func (s *StrategyTable) FindEffectiveStrategyEntry(identifier *component.Identifier) *StrategyTableEntry {
	var PrefixList []string
	for _, v := range identifier.GetComponents() {
		PrefixList = append(PrefixList, v.ToString())
	}
	if v, ok := s.lpm.FindLongestPrefixMatch(PrefixList); ok {
		if strategyTableEntry, ok := v.(*StrategyTableEntry); ok {
			return strategyTableEntry
		}
	}
	return nil
}
