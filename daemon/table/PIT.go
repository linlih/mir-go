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
	"time"
)

/*
PIT表应该设计专门的定时器，用于定时清理超时的PIT表项，PIT提供设置超时回调函数的接口
*/

type PIT struct {
	lpm *LpmMatcher //最长前缀匹配器
}

func CreatePIT() *PIT {
	var p = &PIT{}
	p.lpm = &LpmMatcher{} //初始化
	p.lpm.Create()        //初始化锁
	return p
}

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

func (p *PIT) Insert(interest *packet.Interest, logicFaceId uint64, expireTime time.Duration) *PITEntry {
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
		entry.ExpireTime = expireTime
		entry.InsertOrUpdateInRecord(logicFaceId, interest)
		return entry
	})
	return val.(*PITEntry)
}

// 若为nil 说明 没有找到
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

func (p *PIT) EraseByPITEntry(pitEntry *PITEntry) error {
	var PrefixList []string
	for _, v := range pitEntry.Identifier.GetComponents() {
		PrefixList = append(PrefixList, v.ToString())
	}
	return p.lpm.Delete(PrefixList)
}

func (p *PIT) EraseByLogicFaceID(logicFaceId uint64) uint64 {
	return p.lpm.TraverseFunc(func(val interface{}) uint64 {
		if pitEntry, ok := val.(*PITEntry); ok {
			var ok1, ok2 bool
			if _, ok1 = pitEntry.InRecordList[logicFaceId]; ok1 {
				delete(pitEntry.InRecordList, logicFaceId)
			}
			if _, ok2 = pitEntry.OutRecordList[logicFaceId]; ok2 {
				delete(pitEntry.OutRecordList, logicFaceId)
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
