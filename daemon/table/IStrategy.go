// Package table
// @Author: Jianming Que
// @Description:
// @Version: 1.0.0
// @Date: 2021/3/3 3:21 下午
// @Copyright: MIN-Group；国家重大科技基础设施——未来网络北大实验室；深圳市信息论与未来网络重点实验室
//
package table

import (
	"minlib/packet"
	"mir-go/daemon/lf"
)

type IStrategy interface {
	//////////////////////////////////////////////////////////////////////////////////////////////////////
	//// Triggers
	//////////////////////////////////////////////////////////////////////////////////////////////////////

	// AfterReceiveInterest
	// 当收到一个兴趣包时，会触发本触发器
	//
	// @Description:
	//	Interest 需要满足以下条件：
	//		- Interest 不是回环的
	//		- Interest 没有命中缓存
	//		- Interest 位于当前策略的命名空间下
	//
	// @param ingress		Interest到来的入口LogicFace
	// @param interest		收到的兴趣包
	// @param pitEntry		兴趣包对应的PIT条目
	//
	AfterReceiveInterest(ingress *lf.LogicFace, interest *packet.Interest, pitEntry *PITEntry)

	// AfterContentStoreHit
	// 当兴趣包命中缓存时，会触发本触发器
	//
	// @Description:
	//
	// @param ingress		Interest到来的入口LogicFace
	// @param data			缓存中得到的可以满足 Interest 的 data
	// @param entry			兴趣包对应的PIT条目
	//
	AfterContentStoreHit(ingress *lf.LogicFace, data *packet.Data, entry *PITEntry)

	// AfterReceiveData
	// 当收到一个 data 时，会触发本触发器
	//
	// @Description:
	//	data 应当满足下列条件：
	//		- data 被验证过可以匹配对应的PIT条目
	//		- data 位于当前策略的命名空间下
	// @param ingress		data 到来的入口 LogicFace
	// @param data			收到的 data
	// @param pitEntry		data 对应匹配的PIT条目
	//
	AfterReceiveData(ingress *lf.LogicFace, data *packet.Data, pitEntry *PITEntry)

	// AfterReceiveNack
	// 当收到一个 Nack 时，会触发本触发器
	//
	// @Description:
	//
	// @param ingress		Nack 到来的入口 LogicFace
	// @param nack			收到的 Nack
	// @param pitEntry		Nack 对应匹配的PIT条目
	//
	AfterReceiveNack(ingress *lf.LogicFace, nack *packet.Nack, pitEntry *PITEntry)

	// AfterReceiveGPPkt
	// 当收到一个 GPPkt 时，会触发本触发器
	//
	// @Description:
	// @param ingress		GPPkt 到来的入口 LogicFace
	// @param gPPkt		收到的 GPPkt
	//
	AfterReceiveGPPkt(ingress *lf.LogicFace, gPPkt *packet.GPPkt)

	////////////////////////////////////////////////////////////////////////////////////////////////////////
	////// Actions
	////////////////////////////////////////////////////////////////////////////////////////////////////////
	//
	////
	//// 将 Interest 从指定的逻辑接口转发出去
	////
	//// @Description:
	//// @param egress		转发 Interest 的出口 LogicFace
	//// @param interest		要转发的 Interest
	//// @param entry			Interest 对应匹配的 PIT 条目
	////
	//sendInterest(egress *lf.LogicFace, interest *packet.Interest, pitEntry *PITEntry)
	//
	////
	//// 将 data 从指定的逻辑接口转发出去
	////
	//// @Description:
	//// @param egress		转发 data 的出口 LogicFace
	//// @param data			要转发的 data
	//// @param pitEntry		data 对应匹配的 PIT 条目
	////
	//sendData(egress *lf.LogicFace, data *packet.data, pitEntry *PITEntry)
	//
	////
	//// 将 data 发送给对应 PIT 条目记录的所有符合条件的下游节点
	////
	//// @Description:
	//// @param ingress		data 到来的入口 LogicFace => 主要是用来避免往收到 data 包的 LogicFace 转发 data
	//// @param data			要转发的 data
	//// @param pitEntry		data 对应匹配的 PIT 条目
	////
	//sendDataToAll(ingress *lf.LogicFace, data *packet.data, pitEntry *PITEntry)
	//
	////
	//// 往指定的逻辑接口发送一个 Nack
	////
	//// @Description:
	//// @param egress		转发 Nack 的出口 LogicFace
	//// @param nackHeader	要转发出的Nack的元信息
	//// @param pitEntry		Nack 对应匹配的 PIT 条目
	////
	//sendNack(egress *lf.LogicFace, nackHeader *component.NackHeader, pitEntry *PITEntry)
	//
	////
	//// 将 Nack 发送给对应 PIT 条目记录的所有符合条件的下游节点
	////
	//// @Description:
	//// @param ingress		收到 Nack 的入口 LogicFace
	//// @param nackHeader	要转发出的Nack的元信息
	//// @param pitEntry		Nack 对应匹配的 PIT 条目
	////
	//sendNackToAll(ingress *lf.LogicFace, nackHeader *component.NackHeader, pitEntry *PITEntry)
	//
	////
	//// 往指定的逻辑接口发送一个 GPPkt
	////
	//// @Description:
	//// @param egress		转发 GPPkt 的出口 LogicFace
	//// @param gPPkt		要转发出的 GPPkt
	////
	//sendGPPkt(egress *lf.LogicFace, gPPkt *packet.GPPkt)
	//
	////
	//// 让PIT条目触发立即过期并清除的操作
	////
	//// @Description:
	////  本函数会设置 PIT 条目的超时时间为当前时间，以触发立即超时。
	////  策略模块如果发现兴趣包无法转发到上游，并且不想等待上游节点返回数据时，可以调用本方法
	//// @receiver s
	//// @param pitEntry
	////
	//rejectPendingInterest(pitEntry *PITEntry)
	//
	////////////////////////////////////////////////////////////////////////////////////////////////////////
	////// 其它辅助函数
	////////////////////////////////////////////////////////////////////////////////////////////////////////
	//
	////
	//// 在 FIB 表中查询可用于转发 Interest 的 FIB 条目
	////
	//// @Description:
	//// @param interest
	////
	//lookupFibForInterest(interest *packet.Interest) *FIBEntry
	//
	////
	//// 在 FIB 表中查询可用于转发 GPPkt 的 FIB 条目
	////
	//// @Description:
	//// @param gPPkt
	////
	//lookupFibForGPPkt(gPPkt *packet.GPPkt) *FIBEntry
}
