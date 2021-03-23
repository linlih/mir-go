//
// @Author: weiguohua
// @Description:
// @Version: 1.0.0
// @Date: 2021/3/18 上午9:29
// @Copyright: MIN-Group；国家重大科技基础设施——未来网络北大实验室；深圳市信息论与未来网络重点实验室
//
package lf

import (
	"container/heap"
	"minlib/packet"
	"strconv"
	"time"
)

// 如果超过500ms没收到下一个分片，则删除表项
const ReassembleTimeout = 500 // 500ms超时

//
// @Description: 包重组器
//
type LpReassemble struct {
	// key : string = "remoteMacAddr:packetID"
	mPartialPackets  map[string]*PartialPacket
	timeoutEventHeap TimeoutEventHeap
}

//
// @Description: 初始化包重组器
// @receiver l
//
func (l *LpReassemble) Init() {
	l.mPartialPackets = make(map[string]*PartialPacket)
}

//
// @Description: 处理超时事件，每次ReceiveFragment时调用
// @receiver l
// @param curTime	当前时间戳 ms
//
func (l *LpReassemble) processTimeoutEvent(curTime int64) {

	for l.timeoutEventHeap.Len() > 0 && l.timeoutEventHeap[0].timeoutTime > curTime {
		timeOutEvent := heap.Pop(&l.timeoutEventHeap).(TimeoutEvent)

		if l.mPartialPackets[timeOutEvent.key].dropTime < curTime {
			delete(l.mPartialPackets, timeOutEvent.key)
		} else {
			timeOutEvent.timeoutTime = l.mPartialPackets[timeOutEvent.key].dropTime
			heap.Push(&l.timeoutEventHeap, timeOutEvent)
		}
	}
}

//
// @Description: 对收到的分片进行重组。调用 processTimeoutEvent 处理超时事件，
// @receiver l
// @param remoteMacAddr	对端Mac地址
// @param lpPacket
// @return *packet.LpPacket
//
func (l *LpReassemble) ReceiveFragment(remoteMacAddr string, lpPacket *packet.LpPacket) *packet.LpPacket {
	curTime := time.Now().UnixNano() / 1000000
	key := remoteMacAddr + ":" + strconv.FormatUint(lpPacket.GetId(), 10)
	entry, ok := l.mPartialPackets[key]
	if ok {
		entry.AddLpPacket(lpPacket, curTime)
	} else {
		entry = new(PartialPacket)
		entry.AddLpPacket(lpPacket, curTime)
	}

	reassembleLpPacket := entry.DoReassemble()

	if reassembleLpPacket != nil {
		delete(l.mPartialPackets, key)
	} else {
		l.mPartialPackets[key] = entry
	}
	l.processTimeoutEvent(curTime)
	return reassembleLpPacket
}
