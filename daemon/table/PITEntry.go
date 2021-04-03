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
	"minlib/packet"
	"mir-go/daemon/common"
	"mir-go/daemon/lf"
	"sync"
	"sync/atomic"
	"time"
)

//
// 流入记录表结构体
//
// @Description:流入记录表结构体
//
type InRecord struct {
	LogicFace  *lf.LogicFace    //流入LogicFace指针
	Interest   *packet.Interest //兴趣包指针
	ExpireTime uint64           //超时时间 应用层设置 底层不用
	LastNonce  component.Nonce  //与最后加入记录表的兴趣包中的nonce一致
}

//
// 流出记录表结构体
//
// @Description:流出记录表结构体
//
type OutRecord struct {
	LogicFace  *lf.LogicFace   //流出LogicFace指针
	ExpireTime uint64          //超时时间 应用层设置 底层不用
	LastNonce  component.Nonce //与InRecord中的LastNonce一致
	NackHeader *component.NackHeader
}

//
// PITEntry结构体 PIT表项
//
// @Description:PITEntry结构体 PIT表项 存储在PIT前缀树的节点中
//
type PITEntry struct {
	Identifier *component.Identifier //标识对象指针
	//ExpireTime    time.Duration         //超时时间 底层设置 过期删除
	InRecordList           map[uint64]*InRecord  //流入记录表
	OutRecordList          map[uint64]*OutRecord //流出记录表
	InRWlock               *sync.RWMutex         //流入读写锁
	OutRWlock              *sync.RWMutex         //流出读写锁
	Ticker                 *time.Ticker          //定时器
	ch                     chan int              //取消定时器信号
	isSatisfiedAtomicValue atomic.Value          // 是否已被满足
	isDeletedAtomicValue   atomic.Value          // 是否已经从 PIT 表中移除
}

//
// 初始化PITEntry并返回
//
// @Description:
// @return *PITEntry
//
func CreatePITEntry() *PITEntry {
	var p = &PITEntry{}
	p.InRecordList = make(map[uint64]*InRecord)
	p.OutRecordList = make(map[uint64]*OutRecord)
	p.InRWlock = new(sync.RWMutex)
	p.OutRWlock = new(sync.RWMutex)
	p.ch = make(chan int)
	p.isSatisfiedAtomicValue.Store(false)
	p.isDeletedAtomicValue.Store(false)
	return p
}

//
// 返回当前 PITEntry 是否已经被满足
//
// @Description:
// @receiver p
// @return bool
//
func (p *PITEntry) IsSatisfied() bool {
	return p.isSatisfiedAtomicValue.Load().(bool)
}

//
// 设置当前 PITEntry 是否已经被满足
//
// @Description:
// @receiver p
// @param isSatisfied
//
func (p *PITEntry) SetSatisfied(isSatisfied bool) {
	p.isSatisfiedAtomicValue.Store(isSatisfied)
}

//
// 返回当前 PITEntry 是否已经从 PIT 表中移除
//
// @Description:
// @receiver p
// @return bool
//
func (p *PITEntry) IsDeleted() bool {
	return p.isDeletedAtomicValue.Load().(bool)
}

//
// 设置当前 PITEntry 是否已经从 PIT 表中移除
//
// @Description:
// @receiver p
// @param isDeleted
//
func (p *PITEntry) SetDeleted(isDeleted bool) {
	p.isDeletedAtomicValue.Store(isDeleted)
}

//
// 设置超时定时器 经过duration时间段 自动调用函数f 并且可以在中途调用CancelTimer取消
//
// @Description:
//		p.CancelTimer()
//		time.Sleep(1*time.Millisecond)
//		p.SetExpiryTimer(1*time.Second)
//		加上sleep 不然 上一个函数还没有给Ticker == nil 下一个函数直接执行进入reset函数
// @param time.Duration,func(*PITEntry) 超时时间 和 执行函数
//
func (p *PITEntry) SetExpiryTimer(duration time.Duration, f func(*PITEntry)) {
	if p.Ticker == nil {
		if duration == 0 {
			f(p)
			return
		}
		p.Ticker = time.NewTicker(duration)
		go func() {
			select {
			case <-p.Ticker.C:
				common.LogInfo("执行回调函数")
				f(p)
			case p.ch <- 1:
				common.LogInfo("取消定时器 直接退出")
				p.ch = make(chan int)
			}
			p.Ticker.Stop()
			p.Ticker = nil
		}()
	} else {
		p.Ticker.Reset(duration)
	}
}

//
// 取消定时器
//
// @Description:
//
func (p *PITEntry) CancelTimer() {
	//定时器设置空
	if p.Ticker != nil {
		<-p.ch
		return
	}
	common.LogWarn("the ticker not start")
}

//
// 获得表项中的兴趣包指针 表项中的所有兴趣包都是相同的 但是其他属性不同
//
// @Description:
// @return *packet.Interest, bool
//
func (p *PITEntry) GetInterest() (*packet.Interest, bool) {
	p.InRWlock.RLock()
	defer p.InRWlock.RUnlock()
	for _, inRecord := range p.InRecordList {
		return inRecord.Interest, true
	}
	return nil, false
}

//
// 获得表项中的标识指针
//
// @Description:
// @return *component.Identifier
//
func (p *PITEntry) GetIdentifier() *component.Identifier {
	return p.Identifier
}

//
// 判断表项是否和传入的兴趣包匹配 随机取一个 因为表项中存储的兴趣包都一样
//
// @Description:
// @return bool, error
//
func (p *PITEntry) CanMatch(interest *packet.Interest) (bool, error) {
	p.InRWlock.RLock()
	defer p.InRWlock.RUnlock()
	for _, inRecord := range p.InRecordList {
		return inRecord.Interest.MatchesInterest(interest), nil
	}
	return false, createPITEntryErrorByType(InterestNotExistedError)
}

//
// 获得流入记录列表
//
// @Description:
// @return []InRecord
//
func (p *PITEntry) GetInRecords() []*InRecord {
	InRecordList := make([]*InRecord, 0)
	p.InRWlock.RLock()
	for _, inRecord := range p.InRecordList {
		InRecordList = append(InRecordList, inRecord)
	}
	p.InRWlock.RUnlock()
	return InRecordList
}

//
// 判断流入记录列表是否为空 true 不空 false 空
//
// @Description:
// @return bool
//
func (p *PITEntry) HasInRecords() bool {
	p.InRWlock.RLock()
	defer p.InRWlock.RUnlock()
	return len(p.InRecordList) != 0
}

//
// 根据logicFace从流入记录表中取出对应的流入记录
//
// @Description:
// @return InRecord, error
//
func (p *PITEntry) GetInRecord(logicFace *lf.LogicFace) (*InRecord, error) {
	p.InRWlock.RLock()
	defer p.InRWlock.RUnlock()
	if inRecord, ok := p.InRecordList[logicFace.LogicFaceId]; ok {
		return inRecord, nil
	}
	return &InRecord{}, createPITEntryErrorByType(InRecordNotExistedError)
}

//
// 在PITEntry中插入或更新流入记录
//
// @Description:
// @param uint64,*packet.Interest
// @return *InRecord
//
func (p *PITEntry) InsertOrUpdateInRecord(logicFace *lf.LogicFace, interest *packet.Interest) *InRecord {
	//if p.InRecordList == nil {
	//	p.RWlock.Lock()
	//	p.InRecordList = make(map[uint64]InRecord)
	//	p.RWlock.Unlock()
	//	return &InRecord{}
	//}
	p.InRWlock.Lock()
	delete(p.InRecordList, logicFace.LogicFaceId)
	inRecord := &InRecord{LogicFace: logicFace, Interest: interest, LastNonce: interest.Nonce}
	p.InRecordList[logicFace.LogicFaceId] = inRecord
	p.InRWlock.Unlock()
	// 返回引用 对返回值修改就是对原值修改
	return inRecord
}

//
// 根据logicFace删除PITEntry中的流入记录
//
// @Description:
// @param uint64
// @return error
//
func (p *PITEntry) DeleteInRecord(logicFace *lf.LogicFace) error {
	p.InRWlock.Lock()
	defer p.InRWlock.Unlock()
	if _, ok := p.InRecordList[logicFace.LogicFaceId]; ok {
		delete(p.InRecordList, logicFace.LogicFaceId)
		return nil
	}
	return createPITEntryErrorByType(InRecordNotExistedError)
}

//
// 清空流入记录表
//
// @Description:
//
func (p *PITEntry) ClearInRecords() {
	p.InRWlock.Lock()
	defer p.InRWlock.Unlock()
	p.InRecordList = make(map[uint64]*InRecord)
}

//
// 获得流出记录列表
//
// @Description:
// @return []OutRecord
//
func (p *PITEntry) GetOutRecords() []*OutRecord {
	OutRecordList := make([]*OutRecord, 0)
	p.OutRWlock.RLock()
	for _, outRecord := range p.OutRecordList {
		OutRecordList = append(OutRecordList, outRecord)
	}
	p.OutRWlock.RUnlock()
	return OutRecordList
}

//
// 判断流出记录列表是否为空 true 不空 false 空
//
// @Description:
// @return bool
//
func (p *PITEntry) HasOutRecords() bool {
	p.OutRWlock.RLock()
	defer p.OutRWlock.RUnlock()
	return len(p.OutRecordList) != 0
}

//
// 根据logicFace从流出记录表中取出对应的流出记录
//
// @Description:
// @param logicFace
// @return OutRecord, error
//
func (p *PITEntry) GetOutRecord(logicFace *lf.LogicFace) (*OutRecord, error) {
	p.OutRWlock.RLock()
	defer p.OutRWlock.RUnlock()
	if outRecord, ok := p.OutRecordList[logicFace.LogicFaceId]; ok {
		return outRecord, nil
	}
	return &OutRecord{}, createPITEntryErrorByType(OutRecordNotExistedError)
}

//
// 在PITEntry中插入或更新流出记录
//
// @Description:
// @param uint64,*packet.Interest
// @return *OutRecord
//
func (p *PITEntry) InsertOrUpdateOutRecord(logicFace *lf.LogicFace, interest *packet.Interest) *OutRecord {
	//if p.OutRecordList == nil {
	//	p.OutRecordList = make(map[uint64]OutRecord)
	//}
	p.OutRWlock.Lock()
	delete(p.OutRecordList, logicFace.LogicFaceId)
	outRecord := &OutRecord{LogicFace: logicFace, LastNonce: interest.Nonce}
	p.OutRecordList[logicFace.LogicFaceId] = outRecord
	p.OutRWlock.Unlock()
	return outRecord
}

//
// 根据logicFace删除PITEntry中的流出记录
//
// @Description:
// @param uint64
// @return error
//
func (p *PITEntry) DeleteOutRecord(logicFace *lf.LogicFace) error {
	p.OutRWlock.Lock()
	defer p.OutRWlock.Unlock()
	if _, ok := p.InRecordList[logicFace.LogicFaceId]; ok {
		delete(p.InRecordList, logicFace.LogicFaceId)
		return nil
	}
	return createPITEntryErrorByType(OutRecordNotExistedError)
}

//
// 清空流出记录表
//
// @Description:
//
func (p *PITEntry) ClearOutRecords() {
	p.OutRWlock.Lock()
	defer p.OutRWlock.Unlock()
	p.OutRecordList = make(map[uint64]*OutRecord)
}

/////////////////////////////////////////////////////////////////////////////////////////////////////////
///// 错误处理
/////////////////////////////////////////////////////////////////////////////////////////////////////////

const (
	InRecordNotExistedError = iota
	OutRecordNotExistedError
	InterestNotExistedError
)

type PITEntryError struct {
	msg string
}

func (p PITEntryError) Error() string {
	return fmt.Sprintf("NodeError: %s", p.msg)
}

func createPITEntryErrorByType(errorType int) (err PITEntryError) {
	switch errorType {
	case InRecordNotExistedError:
		err.msg = "the InRecord is not existed"
	case OutRecordNotExistedError:
		err.msg = "the OutRecord is not existed"
	case InterestNotExistedError:
		err.msg = "the Interest is not existed"
	default:
		err.msg = "Unknown error"
	}
	return
}
