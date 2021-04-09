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
)

// 如果超过500ms没收到下一个分片，则删除表项
const ReassembleTimeout = 500 // 500ms超时

//
// @Description: 包重组器
//			包重组器使用一个哈希表和一个超时事件堆实现。
//			mPartialPackets 哈希表
//			哈希表的 key 格式为  "<帧的源MAC地址>:<LpPacket编号>"
//			哈希表的 value 是 一个 PartialPacket 结构体， 结构体中有接收到的分片的数据，接收分片计数，分片总数和超时时间等信息
//			timeoutEventHeap 超时事件堆
//			超时事件堆是一个最小堆，最近要超时的哈希表项会保存在堆顶
//
//			包重组器的实现方案：
//			（1） 通过ReceiveFragment函数接收到一个包分片；
//			（2） 通过包分片的源MAC地址和分片ID，在哈希表中查找有无表项，如果有，则将包放入表项  PartialPacket 结构的分片数组的正确位置；
//			（3） 如果没有表项，则新增一个表项，创建合适长度的包分片数组；
//			（4） 调用 PartialPacket 的 DoReassemble 尝试将所有的分片合并，如果能合并出一个LpPacket，则返回LpPacket对象，
//				并删除哈希表记录，否则返回Nil；
//			（5） 最后调用processTimeoutEvent处理重组已经超时的表项
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
//			代码逻辑：
//			（1） 如果当前超时事件最小堆中存在事件，且最近超时的事件已经超时，则不断将超时的事件从堆顶拿出来处理
//			（2） 通过超时事件对象中保存的事件key值，再次到hash表中确认事件是否真的超时，如果hash表中保存的事件也超时了，
//				则把事件从哈希表中删除，而如果哈希表中的事件没有超时，则更新事件超时时间，重新把事件放回堆中
// @receiver l
// @param curTime	当前时间戳 ms
//
func (l *LpReassemble) processTimeoutEvent(curTime int64) {

	for l.timeoutEventHeap.Len() > 0 && l.timeoutEventHeap[0].timeoutTime < curTime {
		timeOutEvent := heap.Pop(&l.timeoutEventHeap).(TimeoutEvent)
		entry, ok := l.mPartialPackets[timeOutEvent.key]
		if ok {
			if entry.dropTime > curTime {
				delete(l.mPartialPackets, timeOutEvent.key)
			} else {
				timeOutEvent.timeoutTime = entry.dropTime
				heap.Push(&l.timeoutEventHeap, timeOutEvent)
			}
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
	curTime := getTimestampMS()
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
