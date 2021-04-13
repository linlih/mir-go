package table

import (
	"fmt"
	"minlib/component"
	"minlib/packet"
	//"mir-go/daemon/lf"
	"strconv"
	"testing"
)

func TestCSSize(t *testing.T) {
	cs := CreateCS()
	data := &packet.Data{}
	identifier, err := component.CreateIdentifierByString("/min")
	if err != nil {
		fmt.Println(err)
	}
	data.SetName(identifier)
	cs.Insert(data)
	fmt.Println(cs.Size())

	identifier, err = component.CreateIdentifierByString("/min/pku")
	if err != nil {
		fmt.Println(err)
	}
	data.SetName(identifier)
	cs.Insert(data)
	fmt.Println(cs.Size())

	identifier, err = component.CreateIdentifierByString("/min/pku/edu")
	if err != nil {
		fmt.Println(err)
	}
	data.SetName(identifier)
	cs.Insert(data)
	fmt.Println(cs.Size())

	identifier, err = component.CreateIdentifierByString("/min/pku/cn")
	if err != nil {
		fmt.Println(err)
	}
	data.SetName(identifier)
	cs.Insert(data)
	fmt.Println(cs.Size())

	identifier, err = component.CreateIdentifierByString("/min/gdcni15")
	if err != nil {
		fmt.Println(err)
	}
	data.SetName(identifier)
	cs.Insert(data)
	fmt.Println(cs.Size())

	identifier, err = component.CreateIdentifierByString("/min/gdcni15/mir")
	if err != nil {
		fmt.Println(err)
	}
	data.SetName(identifier)
	cs.Insert(data)
	fmt.Println(cs.Size())

	identifier, err = component.CreateIdentifierByString("/min/gdcni15/filegator")
	if err != nil {
		fmt.Println(err)
	}
	data.SetName(identifier)
	cs.Insert(data)
	fmt.Println(cs.Size())

	// 删除一个再测
	cs.EraseByIdentifier(identifier)
	fmt.Println(cs.Size())

	//异常空值 直接报错
	//data=&packet.Data{}
	//cs.Insert(data)
	//fmt.Println(cs.Size())
}

func TestCSInsert(t *testing.T) {
	cs := CreateCS()
	data := &packet.Data{}
	identifier, err := component.CreateIdentifierByString("/min/pku/edu")
	if err != nil {
		fmt.Println(err)
	}

	// 测试 异常值插入
	var PrefixList []string
	data.SetName(&component.Identifier{})
	cs.Insert(data)
	for _, v := range identifier.GetComponents() {
		PrefixList = append(PrefixList, v.ToString())
	}
	fmt.Println(PrefixList)
	fmt.Println(cs.lpm.table)

	//测试 正常插入
	data.SetName(identifier)
	cs.Insert(data)
	for _, v := range identifier.GetComponents() {
		PrefixList = append(PrefixList, v.ToString())
	}
	fmt.Println(PrefixList)
	fmt.Println(cs.lpm.table)
}

func TestCSEraseByIdentifier(t *testing.T) {
	cs := CreateCS()
	identifier, err := component.CreateIdentifierByString("/min/pku/edu")
	if err != nil {
		fmt.Println(err)
	}
	data := &packet.Data{}
	data.SetName(identifier)
	cs.Insert(data)
	// 删除存在数据
	fmt.Println(cs.Size())
	fmt.Println(cs.EraseByIdentifier(identifier))
	fmt.Println(cs.Size())
	// 删除不存在数据
	cs.Insert(data)
	fmt.Println(cs.Size())
	identifier, err = component.CreateIdentifierByString("/min/pku")
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(cs.EraseByIdentifier(identifier))
	fmt.Println(cs.Size())

}

func TestCSFind(t *testing.T) {
	cs := CreateCS()
	identifier, err := component.CreateIdentifierByString("/min")
	if err != nil {
		fmt.Println(err)
	}
	data := &packet.Data{}
	interest := &packet.Interest{}
	interest.SetName(identifier)
	data.SetName(identifier)
	cs.Insert(data)
	// 找到数据
	fmt.Println(cs.Find(interest))

	// 没有找到数据
	interest.SetName(&component.Identifier{})
	fmt.Println(cs.Find(interest))
}

// 速度基本与表项个数呈线性关系
func BenchmarkCSSize(b *testing.B) {
	cs := CreateCS()
	identifier, err := component.CreateIdentifierByString("/min")
	if err != nil {
		fmt.Println(err)
	}
	data := &packet.Data{}
	data.SetName(identifier)
	cs.Insert(data)

	identifier, err = component.CreateIdentifierByString("/min/pku")
	if err != nil {
		fmt.Println(err)
	}
	data.SetName(identifier)
	cs.Insert(data)

	identifier, err = component.CreateIdentifierByString("/min/pku/edu")
	if err != nil {
		fmt.Println(err)
	}
	data.SetName(identifier)
	cs.Insert(data)

	identifier, err = component.CreateIdentifierByString("/min/pku/cn")
	if err != nil {
		fmt.Println(err)
	}
	data.SetName(identifier)
	cs.Insert(data)

	identifier, err = component.CreateIdentifierByString("/min/gdcni15")
	if err != nil {
		fmt.Println(err)
	}
	data.SetName(identifier)
	cs.Insert(data)

	identifier, err = component.CreateIdentifierByString("/min/gdcni15/mir")
	if err != nil {
		fmt.Println(err)
	}
	data.SetName(identifier)
	cs.Insert(data)

	identifier, err = component.CreateIdentifierByString("/min/gdcni15/filegator")
	if err != nil {
		fmt.Println(err)
	}
	data.SetName(identifier)
	cs.Insert(data)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cs.Size()
	}
}

func BenchmarkCSEraseByIdentifier(b *testing.B) {
	cs := CreateCS()
	identifier, err := component.CreateIdentifierByString("/min")
	if err != nil {
		fmt.Println(err)
	}
	data := &packet.Data{}
	data.SetName(identifier)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		cs.Insert(data)
		b.StartTimer()
		cs.EraseByIdentifier(identifier)
	}
}

func BenchmarkCSInsert(b *testing.B) {
	cs := CreateCS()
	//构造一个足够大、足够深的前缀树
	identifierString := "/test"
	for i := 1; i <= 100; i++ {
		for j := 1; j <= 100; j++ {
			identifierString = identifierString + strconv.Itoa(j)
			identifier, err := component.CreateIdentifierByString(identifierString)
			if err != nil {
				fmt.Println(err)
			}

			data := &packet.Data{}
			data.SetName(identifier)
			cs.Insert(data)
		}
		identifierString = identifierString + "/test"
		identifier, err := component.CreateIdentifierByString(identifierString)
		if err != nil {
			fmt.Println(err)
		}
		interest := &packet.Interest{}
		interest.SetName(identifier)
		data := &packet.Data{}
		data.SetName(identifier)
		cs.Insert(data)
	}

	fmt.Println(cs.Size())
	identifier1, err := component.CreateIdentifierByString("/test/test2/test/test")
	if err != nil {
		fmt.Println(err)
	}
	interest := &packet.Interest{}
	interest.SetName(identifier1)
	data1 := &packet.Data{}
	data1.SetName(identifier1)
	cs.Insert(data1)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cs.Insert(data1)
	}
}

func BenchmarkCSFind(b *testing.B) {
	cs := CreateCS()
	identifier, err := component.CreateIdentifierByString("/min")
	if err != nil {
		fmt.Println(err)
	}
	interest := &packet.Interest{}
	interest.SetName(identifier)
	data := &packet.Data{}
	data.SetName(identifier)
	cs.Insert(data)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cs.Find(interest)
	}
}
