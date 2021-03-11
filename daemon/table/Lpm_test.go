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

func TestMatcher(t *testing.T) {
	var m LpmMatcher
	m.Create()
	fmt.Println(m.empty())

	s := make([]string, 0)
	s = append(s, "test")
	test1 := "!11"
	m.AddOrUpdate(s, &test1, nil)
	//m.Delete(s)
	m.AddOrUpdate(s, &test1, nil)

	fmt.Println(m.FindLongestPrefixMatch(s))
	fmt.Println(m.FindExactMatch(s))

	lpm, _ := m.FindLongestPrefixMatch(s)

	ss := lpm.(*string)
	fmt.Println(ss)

	test1 = "1111"
	lpm, _ = m.FindLongestPrefixMatch(s)
	fmt.Println(*lpm.(*string))
	m.Delete(s)
}

func Test_Currency(t *testing.T) {
	var m LpmMatcher
	m.Create()
	fmt.Println(m.empty())

	s := make([]string, 0)
	s = append(s, "test")
	s = append(s, "test1")
	test1 := "!11"
	m.AddOrUpdate(s, &test1, nil)

	var wg sync.WaitGroup
	wg.Add(10000)
	var wg1 sync.WaitGroup
	wg1.Add(10000)
	i := 0
	for i < 10000 {
		go func() {
			defer wg.Done()

			m.AddOrUpdate(s, &test1, nil)
		}()
		i++
	}

	j := 0
	for j < 10000 {
		go func() {
			defer wg.Done()
			m.Delete(s)
		}()
		j++
	}
	wg.Wait()
	wg1.Wait()

}
