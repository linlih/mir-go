package table

import (
	"fmt"
	"minlib/component"
	//"minlib/packet"
	//"mir-go/daemon/lf"
	"testing"
	//"strconv"
)

func TestStrategyTableSize(t *testing.T) {
	strategyTable := CreateStrategyTable()
	var istrategy IStrategy
	strategyName := "strategyName"

	identifier, err := component.CreateIdentifierByString("/min")
	if err != nil {
		fmt.Println(err)
	}
	strategyTable.Insert(identifier, strategyName, istrategy)
	fmt.Println(strategyTable.Size())

	identifier, err = component.CreateIdentifierByString("/min/pku")
	if err != nil {
		fmt.Println(err)
	}
	strategyTable.Insert(identifier, strategyName, istrategy)
	fmt.Println(strategyTable.Size())

	identifier, err = component.CreateIdentifierByString("/min/pku/edu")
	if err != nil {
		fmt.Println(err)
	}
	strategyTable.Insert(identifier, strategyName, istrategy)
	fmt.Println(strategyTable.Size())

	identifier, err = component.CreateIdentifierByString("/min/pku/edu/cn")
	if err != nil {
		fmt.Println(err)
	}
	strategyTable.Insert(identifier, strategyName, istrategy)
	fmt.Println(strategyTable.Size())

	// 删除一个再测
	strategyTable.Erase(identifier)
	fmt.Println(strategyTable.Size())
}

func TestStrategyTableErase(t *testing.T) {
	strategyTable := CreateStrategyTable()
	strategyName := "strategyName"
	var istrategy IStrategy
	identifier, err := component.CreateIdentifierByString("/min")
	if err != nil {
		fmt.Println(err)
	}

	strategyTable.Insert(identifier, strategyName, istrategy)
	// 删除存在数据
	fmt.Println(strategyTable.Size())
	fmt.Println(strategyTable.Erase(identifier))
	fmt.Println(strategyTable.Size())
	// 删除不存在数据
	strategyTable.Insert(identifier, strategyName, istrategy)
	fmt.Println(strategyTable.Size())
	identifier, err = component.CreateIdentifierByString("/min/pku")
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(strategyTable.Erase(identifier))
	fmt.Println(strategyTable.Size())
}

func TestStrategyTableInsert(t *testing.T) {
	strategyTable := CreateStrategyTable()
	var istrategy IStrategy
	strategyName := "strategyName"
	identifier, err := component.CreateIdentifierByString("/min")
	if err != nil {
		fmt.Println(err)
	}

	// 测试 异常值插入
	var PrefixList []string
	strategyTable.Insert(identifier, strategyName, istrategy)
	for _, v := range identifier.GetComponents() {
		PrefixList = append(PrefixList, v.ToString())
	}
	fmt.Println(PrefixList)
	fmt.Println(strategyTable.lpm.table)

	//测试 正常插入
	strategyTable.Insert(identifier, strategyName, istrategy)
	for _, v := range identifier.GetComponents() {
		PrefixList = append(PrefixList, v.ToString())
	}
	fmt.Println(PrefixList)
	fmt.Println(strategyTable.lpm.table)
}

//这个只会写成功的测试用例，没测失败的情况
func TestSetDefaultStrategy(t *testing.T) {
	strategyTable := CreateStrategyTable()
	strategyName := "strategyName"
	strategyTable.SetDefaultStrategy(strategyName)
}

func TestFindEffectiveStrategyEntry(t *testing.T) {
	strategyTable := CreateStrategyTable()
	strategyName := "strategyName"
	var istrategy IStrategy
	identifier, err := component.CreateIdentifierByString("/min")
	if err != nil {
		fmt.Println(err)
	}
	strategyTable.Insert(identifier, strategyName, istrategy)

	strategyName1 := "strategyName1"
	identifier1, err := component.CreateIdentifierByString("/min/pku")
	if err != nil {
		fmt.Println(err)
	}
	strategyTable.Insert(identifier1, strategyName1, istrategy)

	strategyName2 := "strategyName2"
	identifier2, err := component.CreateIdentifierByString("/min/pku/edu")
	if err != nil {
		fmt.Println(err)
	}
	strategyTable.Insert(identifier2, strategyName2, istrategy)

	strategyName3 := "strategyName3"
	identifier3, err := component.CreateIdentifierByString("/min/pku/edu/cn")
	if err != nil {
		fmt.Println(err)
	}
	strategyTable.Insert(identifier3, strategyName3, istrategy)

	//查找"/min/pku"
	fmt.Println(strategyTable.FindEffectiveStrategyEntry(identifier1))
	//查找不存在的前缀
	identifier4, err := component.CreateIdentifierByString("/mis")
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(strategyTable.FindEffectiveStrategyEntry(identifier4))
	//按照最长匹配原则匹配
	identifier4, err = component.CreateIdentifierByString("/min/pku/edu/mis")
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(strategyTable.FindEffectiveStrategyEntry(identifier4))
}

func BenchmarkTableSize(b *testing.B) {
	strategyTable := CreateStrategyTable()
	var istrategy IStrategy
	strategyName := "strategyName"

	identifier, err := component.CreateIdentifierByString("/min")
	if err != nil {
		fmt.Println(err)
	}
	strategyTable.Insert(identifier, strategyName, istrategy)

	identifier, err = component.CreateIdentifierByString("/min/pku")
	if err != nil {
		fmt.Println(err)
	}
	strategyTable.Insert(identifier, strategyName, istrategy)

	identifier, err = component.CreateIdentifierByString("/min/pku/edu")
	if err != nil {
		fmt.Println(err)
	}
	strategyTable.Insert(identifier, strategyName, istrategy)

	identifier, err = component.CreateIdentifierByString("/min/pku/edu/cn")
	if err != nil {
		fmt.Println(err)
	}
	strategyTable.Insert(identifier, strategyName, istrategy)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		strategyTable.Size()
	}
}

func BenchmarkStrategyTableInsert(b *testing.B) {
	strategyTable := CreateStrategyTable()
	var istrategy IStrategy
	strategyName := "strategyName"
	identifier, err := component.CreateIdentifierByString("/min")
	if err != nil {
		fmt.Println(err)
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		strategyTable.Insert(identifier, strategyName, istrategy)
	}

}

func BenchmarkStrategyTableErase(b *testing.B) {
	strategyTable := CreateStrategyTable()
	strategyName := "strategyName"
	var istrategy IStrategy
	identifier, err := component.CreateIdentifierByString("/min")
	if err != nil {
		fmt.Println(err)
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		strategyTable.Insert(identifier, strategyName, istrategy)
		b.StartTimer()
		strategyTable.Erase(identifier)
	}

}

func BenchmarkFindEffectiveStrategyEntry(b *testing.B) {
	strategyTable := CreateStrategyTable()
	strategyName := "strategyName"
	var istrategy IStrategy
	identifier, err := component.CreateIdentifierByString("/min")
	if err != nil {
		fmt.Println(err)
	}
	strategyTable.Insert(identifier, strategyName, istrategy)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		strategyTable.FindEffectiveStrategyEntry(identifier)
	}

}

func BenchmarkSetDefaultStrategy(b *testing.B) {
	strategyTable := CreateStrategyTable()
	strategyName := "strategyName"

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		strategyTable.SetDefaultStrategy(strategyName)
	}
}

/*
//插入、查询性能测试
func BenchmarkStrategyTableInsert(b *testing.B) {
//func BenchmarkFindEffectiveStrategyEntry(b *testing.B) {
	strategyTable := CreateStrategyTable()
	var istrategy IStrategy
	strategyName := "strategyName"
	identifierString := "/test"
	for i := 1; i <= 100; i++ {
		for j := 1; j <= 100; j++ {
			identifierString = identifierString + strconv.Itoa(j)
			identifier, err := component.CreateIdentifierByString(identifierString)
			if err != nil {
				fmt.Println(err)
			}
			strategyTable.Insert(identifier, strategyName, istrategy)
		}
		identifierString = identifierString + "/test"
		identifier, err := component.CreateIdentifierByString(identifierString)
		if err != nil {
			fmt.Println(err)
		}
	}

	b.ReportAllocs()
	b.ResetTimer()
	identifier, err := component.CreateIdentifierByString("/test/4/test/test2")
	if err != nil {
		fmt.Println(err)
	}
	for i := 0; i < b.N; i++ {
		strategyTable.Insert(identifier, strategyName, istrategy)
	}

}
*/
