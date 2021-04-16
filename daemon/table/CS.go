/**
 * @Author: yzy
 * @Description:
 * @Version: 1.0.0
 * @Date: 2021/3/4 上午12:48
 *@Copyright: MIN-Group；国家重大科技基础设施——未来网络北大实验室；深圳市信息论与未来网络重点实验室
 */

package table

import (
	"github.com/sirupsen/logrus"
	common2 "minlib/common"
	"minlib/component"
	"minlib/packet"
)

type CS struct {
	lpm    *LpmMatcher //	最长前缀匹配器
	Hits   uint64      //	命中缓存次数
	Misses uint64      //	没有命中缓存次数
}

func CreateCS() *CS {
	var c = &CS{}
	c.lpm = &LpmMatcher{} //初始化
	c.lpm.Create()        //初始化锁
	return c
}

func (c *CS) Init() {
	c.lpm = &LpmMatcher{} //初始化
	c.lpm.Create()        //初始化锁
}

// Size 获得CS的表项数
func (c *CS) Size() uint64 {
	return c.lpm.TraverseFunc(func(val interface{}) uint64 {
		if _, ok := val.(*CSEntry); ok {
			return 1
		} else {
			common2.LogErrorWithFields(logrus.Fields{
				"value": val,
			}, "CSEntry transform fail")
		}
		return 0
	})
}

// Insert 在CS中添加一个Data包 返回CSEntry表项
func (c *CS) Insert(data *packet.Data) *CSEntry {
	var PrefixList []string
	for _, v := range data.GetName().GetComponents() {
		PrefixList = append(PrefixList, v.ToString())
	}
	val := c.lpm.AddOrUpdate(PrefixList, nil, func(val interface{}) interface{} {
		// not ok 那么val == nil 存入标识
		if _, ok := (val).(*CSEntry); !ok {
			// 存入的表项 不是 *CSEntry类型 或者 为nil
			csEntry := CreateCSEntry()
			val = csEntry
		}
		entry := (val).(*CSEntry)
		entry.Data = data
		return entry
	})
	return val.(*CSEntry)
}

// EraseByIdentifier 通过标识删除CS表中的一个数据包
func (c *CS) EraseByIdentifier(identifier *component.Identifier) error {
	var PrefixList []string
	for _, v := range identifier.GetComponents() {
		PrefixList = append(PrefixList, v.ToString())
	}
	return c.lpm.Delete(PrefixList)
}

// Find 通过兴趣包查询CS表项中的数据包
func (c *CS) Find(interest *packet.Interest) *CSEntry {
	var PrefixList []string
	for _, v := range interest.GetName().GetComponents() {
		PrefixList = append(PrefixList, v.ToString())
	}

	if v, ok := c.lpm.FindExactMatch(PrefixList); ok {
		if csEntry, ok := v.(*CSEntry); ok {
			c.Hits++
			return csEntry
		}
	}
	c.Misses++
	return nil
}
