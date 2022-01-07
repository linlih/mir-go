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
	"minlib/component"
	"sync"
)

type StrategyTableEntry struct {
	Identifier   *component.Identifier // 标识
	StrategyName string                // 策略名称
	IStrategy    IStrategy
	RWlock       *sync.RWMutex
}

func CreateStrategyTableEntry() *StrategyTableEntry {
	var s = &StrategyTableEntry{}
	s.RWlock = new(sync.RWMutex)
	return s
}

// 获取策略名称
func (s *StrategyTableEntry) GetStrategyName() string {
	s.RWlock.RLock()
	defer s.RWlock.RUnlock()
	return s.StrategyName
}

// 设置策略名称
func (s *StrategyTableEntry) SetStrategyName(strategyName string) {
	s.RWlock.Lock()
	defer s.RWlock.Unlock()
	s.StrategyName = strategyName
}

// 获取标识前缀
func (s *StrategyTableEntry) GetPrefix() *component.Identifier {
	s.RWlock.RLock()
	defer s.RWlock.RUnlock()
	return s.Identifier
}

// 设置标识前缀
func (s *StrategyTableEntry) SetPrefix(identifier *component.Identifier) {
	s.RWlock.Lock()
	defer s.RWlock.Unlock()
	s.Identifier = identifier
}

// 获取策略条目对应的策略结构体的指针
func (s *StrategyTableEntry) GetStrategy() IStrategy {
	s.RWlock.RLock()
	defer s.RWlock.RUnlock()
	return s.IStrategy
}

// 设置策略条目对应的策略结构体的指针
func (s *StrategyTableEntry) SetStrategy(istrategy IStrategy) {
	s.RWlock.Lock()
	defer s.RWlock.Unlock()
	s.IStrategy = istrategy
}
