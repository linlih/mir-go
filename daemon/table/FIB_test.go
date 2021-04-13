/**
 * @Author: yzy
 * @Description:
 * @Version: 1.0.0
 * @Date: 2021/3/4 上午2:05
 *@Copyright: MIN-Group；国家重大科技基础设施——未来网络北大实验室；深圳市信息论与未来网络重点实验室
 */

package table

import (
	"fmt"
	"minlib/component"
	"mir-go/daemon/lf"
	"testing"
)

// 单元测试
func TestFIBFindLongestPrefixMatch(t *testing.T) {
	// 测试精确匹配
	fib := CreateFIB()
	identifier, err := component.CreateIdentifierByString("/min/pku/edu")
	if err != nil {
		fmt.Println(err)
	}
	fib.AddOrUpdate(identifier, &lf.LogicFace{LogicFaceId: 1}, 1)
	// 打印结果 &{LogicFaceId:0xc000056140 map[1:{LogicFaceId:1 1}] 0xc00001a258}
	fmt.Println(fib.FindLongestPrefixMatch(identifier))

	// 测试最长前缀匹配 存在
	identifier, err = component.CreateIdentifierByString("/min/pku/edu/cn")
	if err != nil {
		fmt.Println(err)
	}
	// 打印结果 &{LogicFaceId:0xc000056140 map[1:{LogicFaceId:1 1}] 0xc00001a258}
	fmt.Println(fib.FindLongestPrefixMatch(identifier))

	// 测试最长前缀匹配 不存在
	identifier, err = component.CreateIdentifierByString("/min/pku")
	if err != nil {
		fmt.Println(err)
	}
	// 打印结果 <nil>
	fmt.Println(fib.FindLongestPrefixMatch(identifier))

	// 测试异常情况 如标识没有初始化
	// 打印结果 <nil>
	fmt.Println(fib.FindLongestPrefixMatch(&component.Identifier{}))

	// 测试异常情况 加入的标识未初始化
	fib.AddOrUpdate(&component.Identifier{}, &lf.LogicFace{LogicFaceId: 1}, 1)
}

func TestFIBFindExactMatch(t *testing.T) {
	// 测试精确匹配 /min/pku/edu
	fib := CreateFIB()
	identifier, err := component.CreateIdentifierByString("/min/pku/edu")
	if err != nil {
		fmt.Println(err)
	}
	fib.AddOrUpdate(identifier, &lf.LogicFace{LogicFaceId: 1}, 1)
	// 打印结果 &{LogicFaceId:0xc0000d0120 map[1:{LogicFaceId:1 1}] 0xc0000ec060}
	fmt.Println(fib.FindExactMatch(identifier))

	// 测试精确匹配 /min/pku/edu/cn
	identifier, err = component.CreateIdentifierByString("/min/pku/edu/cn")
	if err != nil {
		fmt.Println(err)
	}
	// 打印结果 nil
	fmt.Println(fib.FindExactMatch(identifier))

	// 测试精确匹配 /min/pku
	identifier, err = component.CreateIdentifierByString("/min/pku")
	if err != nil {
		fmt.Println(err)
	}
	// 打印结果 nil
	fmt.Println(fib.FindExactMatch(identifier))

	// 测试精确匹配 /min/pku2
	identifier, err = component.CreateIdentifierByString("/min/pku2")
	if err != nil {
		fmt.Println(err)
	}
	// 打印结果 nil
	fmt.Println(fib.FindExactMatch(identifier))

}

func TestFIBAddOrUpdate(t *testing.T) {
	fib := CreateFIB()
	// 测试异常情况 加入的标识未初始化
	fibEntry := fib.AddOrUpdate(&component.Identifier{}, &lf.LogicFace{LogicFaceId: 1}, 1)
	// 打印结果 &{{} {<nil>} []}
	fmt.Println(fibEntry.Identifier)

	// 测试add
	identifier, err := component.CreateIdentifierByString("/min/pku/edu")
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(fib.AddOrUpdate(identifier, &lf.LogicFace{LogicFaceId: 1}, 1))

	// 测试update
	fmt.Println(fib.AddOrUpdate(identifier, &lf.LogicFace{LogicFaceId: 1}, 0))
}

func TestFIBEraseByIdentifier(t *testing.T) {
	fib := CreateFIB()
	identifier, err := component.CreateIdentifierByString("/min/pku/edu")
	if err != nil {
		fmt.Println(err)
	}
	fib.AddOrUpdate(identifier, &lf.LogicFace{LogicFaceId: 1}, 1)
	// 测试删除成功
	// 打印结果 <nil>
	fmt.Println(fib.EraseByIdentifier(identifier))
	// 测试删除失败
	fib.AddOrUpdate(identifier, &lf.LogicFace{LogicFaceId: 1}, 1)
	identifier, err = component.CreateIdentifierByString("/min/pku/edu/cn")
	if err != nil {
		fmt.Println(err)
	}
	// 打印结果 NodeError: the entry is not existed
	fmt.Println(fib.EraseByIdentifier(identifier))
	identifier, err = component.CreateIdentifierByString("/min/pku")
	if err != nil {
		fmt.Println(err)
	}
	// 打印结果 NodeError: the entry is not existed
	fmt.Println(fib.EraseByIdentifier(identifier))
	identifier, err = component.CreateIdentifierByString("/min/pku2")
	if err != nil {
		fmt.Println(err)
	}
	// 打印结果 NodeError: the entry is not existed
	fmt.Println(fib.EraseByIdentifier(identifier))
}

func TestFIBEraseByFIBEntry(t *testing.T) {
	fib := CreateFIB()
	identifier, err := component.CreateIdentifierByString("/min/pku/edu")
	if err != nil {
		fmt.Println(err)
	}
	fibEntry := fib.AddOrUpdate(identifier, &lf.LogicFace{LogicFaceId: 1}, 1)
	// 测试删除成功
	// 打印结果 <nil>
	fmt.Println(fib.EraseByFIBEntry(fibEntry))
}

func TestFIBRemoveNextHopByFace(t *testing.T) {
	fib := CreateFIB()
	identifier, err := component.CreateIdentifierByString("/min/pku/edu")
	if err != nil {
		fmt.Println(err)
	}
	fib.AddOrUpdate(identifier, &lf.LogicFace{LogicFaceId: 1}, 1)
	identifier, err = component.CreateIdentifierByString("/min/pku/cn")
	if err != nil {
		fmt.Println(err)
	}
	fib.AddOrUpdate(identifier, &lf.LogicFace{LogicFaceId: 0}, 1)
	identifier, err = component.CreateIdentifierByString("/min/gdcni15/mir-go")
	if err != nil {
		fmt.Println(err)
	}
	fib.AddOrUpdate(identifier, &lf.LogicFace{LogicFaceId: 1}, 1)
	identifier, err = component.CreateIdentifierByString("/min/gdcni15/filegator")
	if err != nil {
		fmt.Println(err)
	}
	fib.AddOrUpdate(identifier, &lf.LogicFace{LogicFaceId: 0}, 1)
	// 打印结果 2 0 2 0
	fmt.Println(fib.RemoveNextHopByFace(&lf.LogicFace{LogicFaceId: 0}))
	fmt.Println(fib.RemoveNextHopByFace(&lf.LogicFace{LogicFaceId: 0}))
	fmt.Println(fib.RemoveNextHopByFace(&lf.LogicFace{LogicFaceId: 1}))
	fmt.Println(fib.RemoveNextHopByFace(&lf.LogicFace{LogicFaceId: 1}))
}

func TestFIBSize(t *testing.T) {
	fib := CreateFIB()
	identifier, err := component.CreateIdentifierByString("/min")
	if err != nil {
		fmt.Println(err)
	}
	fib.AddOrUpdate(identifier, &lf.LogicFace{LogicFaceId: 1}, 1)
	fmt.Println(fib.Size())
	identifier, err = component.CreateIdentifierByString("/min/pku")
	if err != nil {
		fmt.Println(err)
	}
	fib.AddOrUpdate(identifier, &lf.LogicFace{LogicFaceId: 1}, 1)
	fmt.Println(fib.Size())
	identifier, err = component.CreateIdentifierByString("/min/pku/edu")
	if err != nil {
		fmt.Println(err)
	}
	fib.AddOrUpdate(identifier, &lf.LogicFace{LogicFaceId: 1}, 1)
	fmt.Println(fib.Size())

	identifier, err = component.CreateIdentifierByString("/min/pku/cn")
	if err != nil {
		fmt.Println(err)
	}
	fib.AddOrUpdate(identifier, &lf.LogicFace{LogicFaceId: 0}, 1)
	fmt.Println(fib.Size())

	identifier, err = component.CreateIdentifierByString("/min/gdcni15/mir-go")
	if err != nil {
		fmt.Println(err)
	}
	fib.AddOrUpdate(identifier, &lf.LogicFace{LogicFaceId: 1}, 1)
	fmt.Println(fib.Size())

	identifier, err = component.CreateIdentifierByString("/min/gdcni15/filegator")
	if err != nil {
		fmt.Println(err)
	}
	fib.AddOrUpdate(identifier, &lf.LogicFace{LogicFaceId: 0}, 1)
	fmt.Println(fib.Size())

}

func TestGetDepth(t *testing.T) {
	fib := CreateFIB()
	identifier, err := component.CreateIdentifierByString("/min")
	if err != nil {
		fmt.Println(err)
	}
	fib.AddOrUpdate(identifier, &lf.LogicFace{LogicFaceId: 1}, 1)

	identifier, err = component.CreateIdentifierByString("/min/pku")
	if err != nil {
		fmt.Println(err)
	}
	fib.AddOrUpdate(identifier, &lf.LogicFace{LogicFaceId: 1}, 1)

	identifier, err = component.CreateIdentifierByString("/min/pku/edu")
	if err != nil {
		fmt.Println(err)
	}
	fib.AddOrUpdate(identifier, &lf.LogicFace{LogicFaceId: 1}, 1)

	identifier, err = component.CreateIdentifierByString("/min/pku/cn")
	if err != nil {
		fmt.Println(err)
	}
	fib.AddOrUpdate(identifier, &lf.LogicFace{LogicFaceId: 0}, 1)

	identifier, err = component.CreateIdentifierByString("/min/gdcni15/mir-go")
	if err != nil {
		fmt.Println(err)
	}
	fib.AddOrUpdate(identifier, &lf.LogicFace{LogicFaceId: 1}, 1)

	identifier, err = component.CreateIdentifierByString("/min/gdcni15/filegator")
	if err != nil {
		fmt.Println(err)
	}
	fib.AddOrUpdate(identifier, &lf.LogicFace{LogicFaceId: 0}, 1)

	identifier, err = component.CreateIdentifierByString("/min/gdcni15/filegator/test/test2")
	if err != nil {
		fmt.Println(err)
	}
	fib.AddOrUpdate(identifier, &lf.LogicFace{LogicFaceId: 0}, 1)
	fmt.Println(fib.GetDepth())
}

// 基准测试
// allocs/op表示每个op(单次迭代)发生了多少个不同的内存分配.
// B/op是每操作分配多少个字节.

func BenchmarkFIBFindLongestPrefixMatch(b *testing.B) {
	// 精确匹配
	fib := CreateFIB()
	identifier, err := component.CreateIdentifierByString("/min/pku/edu")
	if err != nil {
		fmt.Println(err)
	}
	fib.AddOrUpdate(identifier, &lf.LogicFace{LogicFaceId: 1}, 1)
	identifier, err = component.CreateIdentifierByString("/min/pku/edu/cn")
	if err != nil {
		fmt.Println(err)
	}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		fib.FindLongestPrefixMatch(identifier)
	}
}

func BenchmarkFIBFindExactMatch(b *testing.B) {
	fib := CreateFIB()
	identifier, err := component.CreateIdentifierByString("/min/pku/edu")
	if err != nil {
		fmt.Println(err)
	}
	fib.AddOrUpdate(identifier, &lf.LogicFace{LogicFaceId: 1}, 1)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		fib.FindExactMatch(identifier)
	}
}

func BenchmarkFIBAddOrUpdate(b *testing.B) {
	fib := CreateFIB()
	identifier, err := component.CreateIdentifierByString("/min/pku/edu")
	if err != nil {
		fmt.Println(err)
	}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		fib.AddOrUpdate(identifier, &lf.LogicFace{LogicFaceId: 1}, 1)
	}
}

func BenchmarkFIBEraseByIdentifier(b *testing.B) {
	fib := CreateFIB()
	identifier, err := component.CreateIdentifierByString("/min/pku/edu")
	if err != nil {
		fmt.Println(err)
	}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		fib.AddOrUpdate(identifier, &lf.LogicFace{LogicFaceId: 1}, 1)
		b.StartTimer()
		fib.EraseByIdentifier(identifier)
	}
}

func BenchmarkFIBEraseByFIBEntry(b *testing.B) {
	fib := CreateFIB()
	identifier, err := component.CreateIdentifierByString("/min/pku/edu")
	if err != nil {
		fmt.Println(err)
	}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		fibEntry := fib.AddOrUpdate(identifier, &lf.LogicFace{LogicFaceId: 1}, 1)
		b.StartTimer()
		fib.EraseByFIBEntry(fibEntry)
	}
}

// b.StopTimer() 消除add函数添加的额外时间 测试时间60s左右 因为启用了定时器
func BenchmarkFIBRemoveNextHopByFace(b *testing.B) {
	fib := CreateFIB()
	identifier, err := component.CreateIdentifierByString("/min/pku/edu")
	if err != nil {
		fmt.Println(err)
	}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		fib.AddOrUpdate(identifier, &lf.LogicFace{LogicFaceId: 1}, 1)
		b.StartTimer()
		fib.RemoveNextHopByFace(&lf.LogicFace{LogicFaceId: 1})
	}
}

func BenchmarkFIBSize(b *testing.B) {
	fib := CreateFIB()
	identifier, err := component.CreateIdentifierByString("/min")
	if err != nil {
		fmt.Println(err)
	}
	fib.AddOrUpdate(identifier, &lf.LogicFace{LogicFaceId: 1}, 1)

	identifier, err = component.CreateIdentifierByString("/min/pku")
	if err != nil {
		fmt.Println(err)
	}
	fib.AddOrUpdate(identifier, &lf.LogicFace{LogicFaceId: 1}, 1)

	identifier, err = component.CreateIdentifierByString("/min/pku/edu")
	if err != nil {
		fmt.Println(err)
	}
	fib.AddOrUpdate(identifier, &lf.LogicFace{LogicFaceId: 1}, 1)

	identifier, err = component.CreateIdentifierByString("/min/pku/cn")
	if err != nil {
		fmt.Println(err)
	}
	fib.AddOrUpdate(identifier, &lf.LogicFace{LogicFaceId: 0}, 1)

	identifier, err = component.CreateIdentifierByString("/min/gdcni15/mir-go")
	if err != nil {
		fmt.Println(err)
	}
	fib.AddOrUpdate(identifier, &lf.LogicFace{LogicFaceId: 1}, 1)

	identifier, err = component.CreateIdentifierByString("/min/gdcni15/filegator")
	if err != nil {
		fmt.Println(err)
	}
	fib.AddOrUpdate(identifier, &lf.LogicFace{LogicFaceId: 0}, 1)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		fib.Size()
	}
}
