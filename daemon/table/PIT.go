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
	"minlib/packet"
	"mir-go/daemon/lf"
)

//
// PIT表结构体
//
// @Description:PIT表结构体,用前缀树存储表项
//
type PIT struct {
	lpm *LpmMatcher //最长前缀匹配器
}

//
// 初始化PIT表 并返回
//
// @Description:
// @return *PIT
//
func CreatePIT() *PIT {
	var p = &PIT{}
	p.lpm = &LpmMatcher{} //初始化
	p.lpm.Create()        //初始化锁
	return p
}

//
// 返回PIT表中含有数据的表项数
//
// @Description:
// @return uint64
//
func (p *PIT) Size() uint64 {
	return p.lpm.TraverseFunc(func(val interface{}) uint64 {
		if _, ok := val.(*PITEntry); ok {
			return 1
		} else {
			fmt.Println("PITEntry transform fail")
		}
		return 0
	})
}

//
// 通过兴趣包在前缀树中精准匹配查找对应的PITEntry
//
// @Description:
// @param *packet.Interest	需要进行查找的兴趣包
// @return *PITEntry error
//
func (p *PIT) Find(interest *packet.Interest) (*PITEntry, error) {
	var PrefixList []string
	for _, v := range interest.GetName().GetComponents() {
		PrefixList = append(PrefixList, v.ToString())
	}
	if v, ok := p.lpm.FindExactMatch(PrefixList); ok {
		if pitEntry, ok := v.(*PITEntry); ok {
			return pitEntry, nil
		}
	}
	return nil, createPITErrorByType(PITEntryNotExistedError)
}

//
// 在PIT表中插入PITEntry
//
// @Description:
// @param *packet.Interest 兴趣包指针
// @return *PITEntry
//
func (p *PIT) Insert(interest *packet.Interest) *PITEntry {
	var PrefixList []string
	for _, v := range interest.GetName().GetComponents() {
		PrefixList = append(PrefixList, v.ToString())
	}
	val := p.lpm.AddOrUpdate(PrefixList, nil, func(val interface{}) interface{} {
		// not ok 那么val == nil 存入标识
		if _, ok := (val).(*PITEntry); !ok {
			// 存入的表项 不是 *PITEntry类型 或者 为nil
			pitEntry := CreatePITEntry()
			val = pitEntry
		}
		entry := (val).(*PITEntry)
		entry.Identifier = interest.GetName()
		return entry
	})
	return val.(*PITEntry)
}

//
// 根据数据包在PIT表中获取PITEntry 匹配过程应该在PIT表中删除匹配过的表项
//
// @Description:
// @param *packet.Data	数据包指针
// @return *PITEntry
//
func (p *PIT) FindDataMatches(data *packet.Data) *PITEntry {
	var PrefixList []string
	for _, v := range data.GetName().GetComponents() {
		PrefixList = append(PrefixList, v.ToString())
	}
	if v, ok := p.lpm.FindExactMatch(PrefixList); ok {
		p.lpm.Delete(PrefixList)
		return v.(*PITEntry)
	}
	return nil
}

//
// 根据PITEntry删除PIT表中的PITEntry
//
// @Description:
// @param *PITEntry
// @return error
//
func (p *PIT) EraseByPITEntry(pitEntry *PITEntry) error {
	var PrefixList []string
	for _, v := range pitEntry.Identifier.GetComponents() {
		PrefixList = append(PrefixList, v.ToString())
	}
	return p.lpm.Delete(PrefixList)
}

//
// 删除所有以logicFace为流入接口号或流出接口号的表项,返回删除的数量
//
// @Description:
// @param logicFace
// @return uint64
//
func (p *PIT) EraseByLogicFace(logicFace *lf.LogicFace) uint64 {
	return p.lpm.TraverseFunc(func(val interface{}) uint64 {
		if pitEntry, ok := val.(*PITEntry); ok {
			var ok1, ok2 bool
			if _, ok1 = pitEntry.InRecordList[logicFace.LogicFaceId]; ok1 {
				delete(pitEntry.InRecordList, logicFace.LogicFaceId)
			}
			if _, ok2 = pitEntry.OutRecordList[logicFace.LogicFaceId]; ok2 {
				delete(pitEntry.OutRecordList, logicFace.LogicFaceId)
			}
			if ok1 || ok2 {
				return 1
			}
			return 0
		}
		fmt.Println("PITEntry transform fail")
		return 0
	})
}

/////////////////////////////////////////////////////////////////////////////////////////////////////////
///// 错误处理
/////////////////////////////////////////////////////////////////////////////////////////////////////////

const (
	PITEntryNotExistedError = iota
)

type PITError struct {
	msg string
}

func (p PITError) Error() string {
	return fmt.Sprintf("NodeError: %s", p.msg)
}

func createPITErrorByType(errorType int) (err PITEntryError) {
	switch errorType {
	case PITEntryNotExistedError:
		err.msg = "PITEntry not found by interest"
	default:
		err.msg = "Unknown error"
	}
	return
}
