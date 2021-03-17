/**
 * @Author: yzy
 * @Description:
 * @Version: 1.0.0
 * @Date: 2021/3/4 上午12:48
 *@Copyright: MIN-Group；国家重大科技基础设施——未来网络北大实验室；深圳市信息论与未来网络重点实验室
 */

package table

import (
	"fmt"
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

func (s *StrategyTable) Init(){
	s.lpm = &LpmMatcher{} //初始化
	s.lpm.Create()        //初始化锁
}

// 获得StrategyTable的大小
func (s *StrategyTable) Size() uint64 {
	return s.lpm.TraverseFunc(func(val interface{}) uint64 {
		if _, ok := val.(*StrategyTableEntry); ok {
			return 1
		} else {
			fmt.Println("StrategyTableEntry transform fail")
		}
		return 0
	})
}

// 为所有的前缀设置一个默认的策略
func (s *StrategyTable) SetDefaultStrategy(strategyName string) {
	s.lpm.TraverseFunc(func(val interface{}) uint64 {
		if strategyTableEntry, ok := val.(*StrategyTableEntry); ok {
			strategyTableEntry.StrategyName = strategyName
			return 1
		} else {
			fmt.Println("StrategyTableEntry transform fail")
		}
		return 0
	})
}

// 往策略表中插入一个策略
func (s *StrategyTable) Insert(identifier *component.Identifier, strategyName string, istrategy IStrategy) *StrategyTableEntry {
	var PrefixList []string
	for _, v := range identifier.GetComponents() {
		PrefixList = append(PrefixList, v.ToString())
	}
	val := s.lpm.AddOrUpdate(PrefixList, nil, func(val interface{}) interface{} {
		// not ok 那么val == nil 存入标识
		if _, ok := (val).(*StrategyTableEntry); !ok {
			// 存入的表项 不是 *StrategyTableEntry类型 或者 为nil
			csEntry := CreateStrategyTableEntry()
			val = csEntry
		}
		entry := (val).(*StrategyTableEntry)
		entry.StrategyName = strategyName
		entry.IStrategy = istrategy
		return entry
	})
	return val.(*StrategyTableEntry)

}

// 通过前缀删除策略表中策略
func (s *StrategyTable) Erase(identifier *component.Identifier) error {
	var PrefixList []string
	for _, v := range identifier.GetComponents() {
		PrefixList = append(PrefixList, v.ToString())
	}
	return s.lpm.Delete(PrefixList)
}

// 查询和一个指定的名称前缀匹配的策略条目 最长前缀匹配
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
