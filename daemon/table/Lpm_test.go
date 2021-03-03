/**
 * @Author: wzx
 * @Description:
 * @Version: 1.0.0
 * @Date: 2021/3/2 上午1:42
 *@Copyright: MIN-Group；国家重大科技基础设施——未来网络北大实验室；深圳市信息论与未来网络重点实验室
 */

package table

import (
	"fmt"
	"sync"
	"testing"
)

func TestMatcher(t *testing.T){
	var m LpmMatcher
	m.Create()
	fmt.Println(m.empty())

	s := make([]string, 0)
	s = append(s, "test")
	test1 := "!11"
	m.AddOrUpdate(s, &test1)
	m.Delete(s)
	m.AddOrUpdate(s, &test1)

	s = append(s, "test1")

	fmt.Println(m.FindExactMatch(s))
	fmt.Println(m.FindLongestPrefixMatch(s))


	lpm,_ := m.FindLongestPrefixMatch(s)

	ss := lpm.(*string)
	fmt.Println(ss)

	test1 = "1111"
	lpm,_ = m.FindLongestPrefixMatch(s)
	fmt.Println(*lpm.(*string))
	m.Delete(s)
}

func Test_Currency(t *testing.T){
	var m LpmMatcher
	m.Create()
	fmt.Println(m.empty())

	s := make([]string, 0)
	s = append(s, "test")
	test1 := "!11"
	m.AddOrUpdate(s, &test1)

	var wg sync.WaitGroup
	wg.Add(1000)
	var wg1 sync.WaitGroup
	wg1.Add(1000)
	count := 0
	i := 0
	for i < 1000 {
		go func() {
			defer wg.Done()
			count++
			m.AddOrUpdate(s, &test1)
		}()
		i++
	}

	j := 0
	for j < 1000 {
		go func() {
			defer wg.Done()
			count++
			m.Delete(s)
		}()

	}
	wg.Wait()
	wg1.Wait()
	fmt.Println(count)
}