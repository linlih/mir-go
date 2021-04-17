// Package fw
// @Author: Jianming Que
// @Description:
// @Version: 1.0.0
// @Date: 2021/3/15 8:58 下午
// @Copyright: MIN-Group；国家重大科技基础设施——未来网络北大实验室；深圳市信息论与未来网络重点实验室
//
package fw

import (
	"github.com/sirupsen/logrus"
	common2 "minlib/common"
	"minlib/component"
	"minlib/packet"
	"mir-go/daemon/lf"
	"mir-go/daemon/table"
)

// BestRouteStrategy
// 最佳路由转发策略实现
//
// @Description:
//
type BestRouteStrategy struct {
	StrategyBase
}

//
// 找到所有可用下一跳中开销最小的下一跳
//
// @Description:
// @receiver brs
// @param ingress
// @param fibEntry
// @return *table.NextHop
//
func (brs *BestRouteStrategy) findLowestCostNextHop(ingress *lf.LogicFace, fibEntry *table.FIBEntry) *table.NextHop {
	// 找到 Cost 最小的下一跳
	var miniHop *table.NextHop = nil
	if fibEntry != nil {
		for _, nextHop := range fibEntry.GetNextHops() {
			// 找到 Cost 最小，并且排除 Interest 到来的逻辑接口
			if (miniHop == nil || miniHop.Cost > nextHop.Cost) && nextHop.LogicFace.LogicFaceId != ingress.LogicFaceId {
				miniHop = nextHop
			}
		}
	}
	return miniHop
}

func (brs *BestRouteStrategy) AfterReceiveInterest(ingress *lf.LogicFace, interest *packet.Interest, pitEntry *table.PITEntry) {
	// 首先判断是否有正在pending的 out-record
	if HasPendingOutRecords(pitEntry) {
		// 不是新的 Interest ，不转发被聚合
		common2.LogDebugWithFields(logrus.Fields{
			"ingress":  ingress.LogicFaceId,
			"interest": interest.ToUri(),
			"pitEntry": pitEntry.Identifier.ToUri(),
		}, "PITEntry already has pending interest, drop")
		return
	}

	// 尝试找到可用的下一跳进行转发
	fibEntry := brs.lookupFibForInterest(interest)

	// 找到开销最小的下一跳
	miniHop := brs.findLowestCostNextHop(ingress, fibEntry)

	if miniHop == nil {
		// 如果没有找到下一跳路由信息，直接返回一个原因为 no-route 的 Nack
		var nh component.NackHeader
		nh.SetNackReason(component.NackReasonNoRoute)
		brs.sendNack(ingress, &nh, pitEntry)

		// 同时触发 PITEntry 移除
		brs.rejectPendingInterest(pitEntry)
		return
	}

	// 将兴趣包转发到可用的下一跳
	brs.sendInterest(miniHop.LogicFace, interest, pitEntry)
}

func (brs *BestRouteStrategy) AfterReceiveNack(ingress *lf.LogicFace, nack *packet.Nack, pitEntry *table.PITEntry) {
	// 保存最不严重的 Nack 原因
	leastSevereReason := component.NackReasonUnknown
	// 保存没有被 Nack 的出记录的数量
	notNackedOutRecordNums := 0

	// 在 Strategy.AfterReceiveNack 之前，OnIncomingNack 管道已经将 Nack 头部信息保存到了对应的 out-record 里面
	// 所以下面的步骤肯定至少能从一个 out-record 中得到 NackHeader
	for _, outRecord := range pitEntry.GetOutRecords() {
		if outRecord.NackHeader != nil {
			if int(outRecord.NackHeader.GetNackReason()) > leastSevereReason {
				leastSevereReason = int(outRecord.NackHeader.GetNackReason())
			}
		} else {
			notNackedOutRecordNums++
		}
	}

	// 如果还有 out-record 没有被 Nack，则不转发nack，等待其它上游返回的nack
	if notNackedOutRecordNums > 0 {
		return
	}

	var nh component.NackHeader
	nh.SetNackReason(uint64(leastSevereReason))

	brs.sendNackToAll(ingress, &nh, pitEntry)
}

func (brs *BestRouteStrategy) AfterReceiveCPacket(ingress *lf.LogicFace, cPacket *packet.CPacket) {
	fibEntry := brs.lookupFibForCPacket(cPacket)
	miniHop := brs.findLowestCostNextHop(ingress, fibEntry)
	if miniHop == nil {
		// 没有路由无法转发
		return
	}
	brs.sendCPacket(miniHop.LogicFace, cPacket)
}
