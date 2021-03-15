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
	"sync"
)

type StrategyTableEntry struct {
	Identifier   *component.Identifier // 标识
	StrategyName string                // 策略名称
	IStrategy    *IStrategy
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
func (s *StrategyTableEntry) GetStrategy() *IStrategy {
	s.RWlock.RLock()
	defer s.RWlock.RUnlock()
	return s.IStrategy
}

// 设置策略条目对应的策略结构体的指针
func (s *StrategyTableEntry) SetStrategy(istrategy *IStrategy) {
	s.RWlock.Lock()
	defer s.RWlock.Unlock()
	s.IStrategy = istrategy
}
