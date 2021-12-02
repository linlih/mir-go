// Package fw
// @Author: Jianming Que
// @Description:
// @Version: 1.0.0
// @Date: 2021/2/22 12:03 下午
// @Copyright: MIN-Group；国家重大科技基础设施——未来网络北大实验室；深圳市信息论与未来网络重点实验室
//
package fw

import (
	"minlib/component"
	"minlib/packet"
	"mir-go/daemon/lf"
	"mir-go/daemon/table"
)

type IForwarder interface {
	// OnIncomingInterest
	// 处理一个兴趣包到来 （ Incoming Interest Pipeline）
	//
	// @Description:
	// @param ingress	入口Face
	// @param interest	收到的内容兴趣包
	//
	OnIncomingInterest(ingress *lf.LogicFace, interest *packet.Interest)

	// OnInterestLoop
	// 处理一个回环的兴趣包 （ Interest Loop Pipeline ）
	//
	// @Description:
	// @param ingress
	// @param interest
	//
	OnInterestLoop(ingress *lf.LogicFace, interest *packet.Interest)

	// OnContentStoreMiss
	// 处理兴趣包未命中缓存 （ ContentStore Miss Pipeline ）
	//
	// @Description:
	// @param ingress
	// @param pitEntry
	// @param interest
	//
	OnContentStoreMiss(ingress *lf.LogicFace, pitEntry *table.PITEntry, interest *packet.Interest)

	// OnContentStoreHit
	// 处理兴趣包命中缓存 （ ContentStore Hit Pipeline ）
	//
	// @Description:
	// @param ingress
	// @param pitEntry
	// @param interest
	// @param data
	//
	OnContentStoreHit(ingress *lf.LogicFace, pitEntry *table.PITEntry, interest *packet.Interest, data *table.CSEntry)

	// OnOutgoingInterest
	// 处理将兴趣包通过 LogicFace 发出 （ Outgoing Interest Pipeline ）
	//
	// @Description:
	// @param egress
	// @param pitEntry
	// @param interest
	//
	OnOutgoingInterest(egress *lf.LogicFace, pitEntry *table.PITEntry, interest *packet.Interest)

	// OnInterestFinalize
	// 兴趣包最终回收处理，此时兴趣包要么被满足要么被Nack （ Interest Finalize Pipeline ）
	//
	// @Description:
	// @param pitEntry
	//
	OnInterestFinalize(pitEntry *table.PITEntry)

	// OnIncomingData
	// 处理一个数据包到来（ Incoming data Pipeline ）
	//
	// @Description:
	// @param ingress
	// @param data
	//
	OnIncomingData(ingress *lf.LogicFace, data *packet.Data)

	// OnDataUnsolicited
	// 收到一个数据包，但是这个数据包是未被请求的 （ data Unsolicited Pipeline ）
	//
	// @Description:
	// @param ingress
	// @param data
	//
	OnDataUnsolicited(ingress *lf.LogicFace, data *packet.Data)

	// OnOutgoingData
	// 处理将一个数据包发出 （ Outgoing data Pipeline ）
	//
	// @Description:
	// @param egress
	// @param data
	//
	OnOutgoingData(egress *lf.LogicFace, data *packet.Data)

	// OnIncomingNack
	// 处理一个 Nack 到来 （ Incoming Nack Pipeline ）
	//
	// @Description:
	// @param ingress
	// @param nack
	//
	OnIncomingNack(ingress *lf.LogicFace, nack *packet.Nack)

	// OnOutgoingNack
	// 处理一个 Nack 发出 （ Outgoing Nack Pipeline ）
	//
	// @Description:
	// @param egress
	// @param pitEntry
	// @param header
	//
	OnOutgoingNack(egress *lf.LogicFace, pitEntry *table.PITEntry, header *component.NackHeader)

	// OnIncomingGPPkt
	// 处理一个 GPPkt 到来 （Incoming GPPkt Pipeline）
	//
	// @Description:
	// @param ingress
	// @param gPPkt
	//
	OnIncomingGPPkt(ingress *lf.LogicFace, gPPkt *packet.GPPkt)

	// OnOutgoingGPPkt
	// 处理一个 GPPkt 发出 （Outgoing GPPkt Pipeline）
	//
	// @Description:
	// @param egress
	// @param gPPkt
	//
	OnOutgoingGPPkt(egress *lf.LogicFace, gPPkt *packet.GPPkt)
}
