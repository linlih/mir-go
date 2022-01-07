// Copyright [2022] [MIN-Group -- Peking University Shenzhen Graduate School Multi-Identifier Network Development Group]
//
// Licensed under the Apache License, Version 2.0 (the "License"): you may
// not use this file except in compliance with the License. You may obtain
// a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS, WITHOUT
// WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the
// License for the specific language governing permissions and limitations
// under the License.

// Package fw
// @Author: Jianming Que
// @Description:
// @Version: 1.0.0
// @Date: 2021/3/4 10:18 上午
// @Copyright: MIN-Group；国家重大科技基础设施——未来网络北大实验室；深圳市信息论与未来网络重点实验室
//
package fw

import (
	"github.com/sirupsen/logrus"
	common2 "minlib/common"
	"minlib/component"
	"minlib/packet"
	"mir-go/daemon/common"
	"mir-go/daemon/lf"
	"mir-go/daemon/table"
)

// StrategyBase
// 本类包含策略类的一个基准实现，其它策略类可以内嵌本对象，并按需覆盖需要使用的回调即可
//
// @Description:
//
type StrategyBase struct {
	forwarder *Forwarder
}

// SetForwarder
// 保存Forwarder指针
//
// @Description:
// @receiver s
// @param forwarder
//
func (s *StrategyBase) SetForwarder(forwarder *Forwarder) {
	s.forwarder = forwarder
}

//////////////////////////////////////////////////////////////////////////////////////////////////////
//// Triggers
//////////////////////////////////////////////////////////////////////////////////////////////////////

// AfterReceiveInterest
// 当收到一个兴趣包时，会触发本触发器（需要子类实现）
//
// @Description:
//	Interest 需要满足以下条件：
//		- Interest 不是回环的
//		- Interest 没有命中缓存
//		- Interest 位于当前策略的命名空间下
//  当本触发器被触发后，策略程序需要决定将 Interest 转发往何处（即从哪个或哪些 LogicFace 将 Interest 转发出去）。大多数策略此时的行为都是通
//  过查询FIB表决定如何转发 Interest ，这个可以通过调用 Strategy.lookupFib 来实现。
//   - 如果策略决定转发该 Interest ，则应该至少调用一次 Strategy.sendInterest 操作将其转发出去；
//   - 如果策略决定不转发该 Interest ，则应该调用 Strategy.setExpiryTimer 操作并将对应PIT条目的超时时间设置为当前时间，使得对应的PIT条目
//     记录可以正确的清除。
//
// @param ingress		Interest到来的入口LogicFace
// @param interest		收到的兴趣包
// @param pitEntry		兴趣包对应的PIT条目
//
func (s *StrategyBase) AfterReceiveInterest(ingress *lf.LogicFace, interest *packet.Interest, pitEntry *table.PITEntry) {
	// 收到一个兴趣包
}

// AfterContentStoreHit
// 当兴趣包命中缓存时，会触发本触发器
//
// @Description:
//  此触发器默认使用 Strategy.sendData 操作将匹配的 data 发送到兴趣包到来方向的下游路由器。
// @param ingress		Interest到来的入口LogicFace
// @param data			缓存中得到的可以满足 Interest 的 data
// @param entry			兴趣包对应的PIT条目
//
func (s *StrategyBase) AfterContentStoreHit(ingress *lf.LogicFace, data *packet.Data, entry *table.PITEntry) {
	common2.LogDebugWithFields(logrus.Fields{
		"ingress":  ingress.LogicFaceId,
		"data":     data.ToUri(),
		"pitEntry": entry.GetIdentifier().ToUri(),
	}, "After content store hit")

	// 命中缓存时直接往兴趣包到来的接口发送一个匹配的 data
	s.sendData(ingress, data, entry)
}

// AfterReceiveData
// 当收到一个 data 时，会触发本触发器
//
// @Description:
//	data 应当满足下列条件：
//		- data 被验证过可以匹配对应的PIT条目
//		- data 位于当前策略的命名空间下
//  此触发器内部应当完成以下功能：
//   - 策略应当通过 Strategy.sendData 或者 Strategy.sendDataToAll 将 data 发送给下游的节点；
//   - 策略可以对 data 进行适当的更改，只要修改之后 data 仍然能够匹配对应的 PIT 条目即可，例如添加或者删除拥塞标记；
//   - 策略应当至少调用一次 Strategy.setExpiryTimer：
//     - 默认情况下， Strategy.setExpiryTimer 将PIT条目的超时时间设置为当前时间，以启动 PIT 条目的清理流程；
//     - 策略也可以选择调用 Strategy.setExpiryTimer 延长 PIT 条目的存活期，从而延迟 data 的转发，只要保证在 PIT 条目被移除之前转发
//       data 即可。
//   - 策略可以在此触发器内收集有关上游的度量信息（比如可以计算RTT）；
//   - 策略可以通过延长收到 data 的PIT条目的生存期，从而等待其它上游节点返回 data （可以从多个上游节点收集 data ，并决策将哪个 data 转发
//     到下游），需要注意的是，对于每一个下有节点，要保证只有一个 data 转发到下游路由器。
// @param ingress		data 到来的入口 LogicFace
// @param data			收到的 data
// @param pitEntry		data 对应匹配的PIT条目
//
func (s *StrategyBase) AfterReceiveData(ingress *lf.LogicFace, data *packet.Data, pitEntry *table.PITEntry) {
	common2.LogDebugWithFields(logrus.Fields{
		"ingress":  ingress.LogicFaceId,
		"data":     data.ToUri(),
		"pitEntry": pitEntry.GetIdentifier().ToUri(),
	}, "After receive data")
	s.sendDataToAll(ingress, data, pitEntry)
}

// AfterReceiveNack
// 当收到一个 Nack 时，会触发本触发器（默认不做任何处理）
//
// @Description:
//  当 After Receive Nack 触发器被触发后，策略程序通常可以执行下述的某一种操作：
//   - 通过调用 send Interest 操作将其转发到相同或不同的上游来重试兴趣（ Retry the Interest ）。大多数策略都需要一个FIB条目来找出潜在的
//     上游，这可以通过调用 Strategy.lookupFib 访问器函数获得；
//   - 通过调用 send Nack 操作将 Nack 反回到下游，放弃对该 Interest 的重传尝试；
//   - 不对这个 Nack 做任何处理。如果 Nack 对应的 Interest 转发给了多个上游，且某些（但不是全部）上游回复了 Nack ，则该策略可能要等待来自
//     更多上游的 data 或 Nack 。
// @param ingress		Nack 到来的入口 LogicFace
// @param nack			收到的 Nack
// @param pitEntry		Nack 对应匹配的PIT条目
//
func (s *StrategyBase) AfterReceiveNack(ingress *lf.LogicFace, nack *packet.Nack, pitEntry *table.PITEntry) {
	common2.LogDebug(logrus.Fields{
		"ingress":  ingress.LogicFaceId,
		"nack":     nack.Interest.ToUri(),
		"reason":   nack.GetNackReason(),
		"pitEntry": pitEntry.GetIdentifier().ToUri(),
	}, "After receive nack")
}

// AfterReceiveGPPkt
// 当收到一个 GPPkt 时，会触发本触发器（需要子类实现）
//
// @Description:
//  当 After Receive GPPkt 触发器被触发后，策略程序通常的行为为查询FIB表，找到可用的路由将 GPPkt 转发出去
// @param ingress		GPPkt 到来的入口 LogicFace
// @param gPPkt		收到的 GPPkt
//
func (s *StrategyBase) AfterReceiveGPPkt(ingress *lf.LogicFace, gPPkt *packet.GPPkt) {
	panic("implement me")
}

//////////////////////////////////////////////////////////////////////////////////////////////////////
//// Actions
//////////////////////////////////////////////////////////////////////////////////////////////////////

//
// 将 Interest 从指定的逻辑接口转发出去
//
// @Description:
// @param egress		转发 Interest 的出口 LogicFace
// @param interest		要转发的 Interest
// @param entry			Interest 对应匹配的 PIT 条目
//
func (s *StrategyBase) sendInterest(egress *lf.LogicFace, interest *packet.Interest, pitEntry *table.PITEntry) {
	s.forwarder.OnOutgoingInterest(egress, pitEntry, interest)
}

//
// 将 data 从指定的逻辑接口转发出去
//
// @Description:
// @param egress		转发 data 的出口 LogicFace
// @param data			要转发的 data
// @param pitEntry		data 对应匹配的 PIT 条目
//
func (s *StrategyBase) sendData(egress *lf.LogicFace, data *packet.Data, pitEntry *table.PITEntry) {
	// data 发出后，对应的入记录应该删除
	if err := pitEntry.DeleteInRecord(egress); err != nil {
		common2.LogErrorWithFields(logrus.Fields{
			"egress":   egress.LogicFaceId,
			"data":     data.ToUri(),
			"pitEntry": pitEntry.GetIdentifier().ToUri(),
		}, "Strategy sendData => delete in-record failed")
	}
	s.forwarder.OnOutgoingData(egress, data)
}

//
// 将 data 发送给对应 PIT 条目记录的所有符合条件的下游节点
//
// @Description:
// @param ingress		data 到来的入口 LogicFace => 主要是用来避免往收到 data 包的 LogicFace 转发 data
// @param data			要转发的 data
// @param pitEntry		data 对应匹配的 PIT 条目
//
func (s *StrategyBase) sendDataToAll(ingress *lf.LogicFace, data *packet.Data, pitEntry *table.PITEntry) {
	now := common.GetCurrentTime()
	downStreams := make([]*lf.LogicFace, 0)

	// 找到所有还没有过期，且不是 data 到来的下游，并向所有符合条件的下游转发一个 data 的备份
	for _, inRecord := range pitEntry.GetInRecords() {
		if inRecord.ExpireTime > now && inRecord.LogicFace.LogicFaceId != ingress.LogicFaceId {
			downStreams = append(downStreams, inRecord.LogicFace)
			// 不能在循环里面直接调用 sendData，因为 sendData 中有删除 in-record 的操作
		}
	}

	for _, downStream := range downStreams {
		s.sendData(downStream, data, pitEntry)
	}
}

//
// 往指定的逻辑接口发送一个 Nack
//
// @Description:
// @param egress		转发 Nack 的出口 LogicFace
// @param nackHeader	要转发出的Nack的元信息
// @param pitEntry		Nack 对应匹配的 PIT 条目
//
func (s *StrategyBase) sendNack(egress *lf.LogicFace, nackHeader *component.NackHeader, pitEntry *table.PITEntry) {
	s.forwarder.OnOutgoingNack(egress, pitEntry, nackHeader)
}

//
// 将 Nack 发送给对应 PIT 条目记录的所有符合条件的下游节点
//
// @Description:
// @param ingress		收到 Nack 的入口 LogicFace
// @param nackHeader	要转发出的Nack的元信息
// @param pitEntry		Nack 对应匹配的 PIT 条目
//
func (s *StrategyBase) sendNackToAll(ingress *lf.LogicFace, nackHeader *component.NackHeader, pitEntry *table.PITEntry) {
	downStreams := make([]*lf.LogicFace, len(pitEntry.GetInRecords()))
	for index, inRecord := range pitEntry.GetInRecords() {
		if inRecord.LogicFace.LogicFaceId != ingress.LogicFaceId {
			downStreams[index] = inRecord.LogicFace
		}
	}
	for _, downStream := range downStreams {
		s.sendNack(downStream, nackHeader, pitEntry)
	}
}

//
// 往指定的逻辑接口发送一个 GPPkt
//
// @Description:
// @param egress		转发 GPPkt 的出口 LogicFace
// @param gPPkt		要转发出的 GPPkt
//
func (s *StrategyBase) sendGPPkt(egress *lf.LogicFace, gPPkt *packet.GPPkt) {
	s.forwarder.OnOutgoingGPPkt(egress, gPPkt)
}

//
// 让PIT条目触发立即过期并清除的操作
//
// @Description:
//  本函数会设置 PIT 条目的超时时间为当前时间，以触发立即超时。
//  策略模块如果发现兴趣包无法转发到上游，并且不想等待上游节点返回数据时，可以调用本方法
// @receiver s
// @param pitEntry
//
func (s *StrategyBase) rejectPendingInterest(pitEntry *table.PITEntry) {
	s.forwarder.SetExpiryTime(pitEntry, 0)
}

//////////////////////////////////////////////////////////////////////////////////////////////////////
//// 其它辅助函数
//////////////////////////////////////////////////////////////////////////////////////////////////////

//
// 在 FIB 表中查询可用于转发 Interest 的 FIB 条目
//
// @Description:
// @param interest
//
func (s *StrategyBase) lookupFibForInterest(interest *packet.Interest) *table.FIBEntry {
	return s.forwarder.FIB.FindLongestPrefixMatch(interest.GetName())
}

//
// 在 FIB 表中查询可用于转发 GPPkt 的 FIB 条目
//
// @Description:
// @param gPPkt
//
func (s *StrategyBase) lookupFibForGPPkt(gPPkt *packet.GPPkt) *table.FIBEntry {
	return s.forwarder.FIB.FindLongestPrefixMatch(gPPkt.DstIdentifier())
}
