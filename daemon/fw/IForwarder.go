//
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
	"mir/daemon/lf"
	"mir/daemon/table"
)

type IForwarder interface {
	//
	// 处理一个兴趣包到来 （ Incoming Interest Pipeline）
	//
	// @Description:
	// @param ingress	入口Face
	// @param interest	收到的内容兴趣包
	//
	OnIncomingInterest(ingress *lf.LogicFace, interest *packet.Interest)

	//
	// 处理一个回环的兴趣包 （ Interest Loop Pipeline ）
	//
	// @Description:
	// @param ingress
	// @param interest
	//
	OnInterestLoop(ingress *lf.LogicFace, interest *packet.Interest)

	//
	// 处理兴趣包未命中缓存 （ ContentStore Miss Pipeline ）
	//
	// @Description:
	// @param ingress
	// @param pitEntry
	// @param interest
	//
	OnContentStoreMiss(ingress *lf.LogicFace, pitEntry *table.PITEntry, interest *packet.Interest)

	//
	// 处理兴趣包命中缓存 （ ContentStore Hit Pipeline ）
	//
	// @Description:
	// @param ingress
	// @param pitEntry
	// @param interest
	// @param data
	//
	OnContentStoreHit(ingress *lf.LogicFace, pitEntry *table.PITEntry, interest *packet.Interest, data *packet.Data)

	//
	// 处理将兴趣包通过 LogicFace 发出 （ Outgoing Interest Pipeline ）
	//
	// @Description:
	// @param egress
	// @param pitEntry
	// @param interest
	//
	OnOutgoingInterest(egress *lf.LogicFace, pitEntry *table.PITEntry, interest *packet.Interest)

	//
	// 兴趣包最终回收处理，此时兴趣包要么被满足要么被Nack （ Interest Finalize Pipeline ）
	//
	// @Description:
	// @param pitEntry
	//
	OnInterestFinalize(pitEntry *table.PITEntry)

	//
	// 处理一个数据包到来（ Incoming Data Pipeline ）
	//
	// @Description:
	// @param ingress
	// @param data
	//
	OnIncomingData(ingress *lf.LogicFace, data *packet.Data)

	//
	// 收到一个数据包，但是这个数据包是未被请求的 （ Data Unsolicited Pipeline ）
	//
	// @Description:
	// @param ingress
	// @param data
	//
	OnDataUnsolicited(ingress *lf.LogicFace, data *packet.Data)

	//
	// 处理将一个数据包发出 （ Outgoing Data Pipeline ）
	//
	// @Description:
	// @param egress
	// @param data
	//
	OnOutgoingData(egress *lf.LogicFace, data *packet.Data)

	//
	// 处理一个 Nack 到来 （ Incoming Nack Pipeline ）
	//
	// @Description:
	// @param ingress
	// @param nack
	//
	OnIncomingNack(ingress *lf.LogicFace, nack *packet.Nack)

	//
	// 处理一个 Nack 发出 （ Outgoing Nack Pipeline ）
	//
	// @Description:
	// @param egress
	// @param pitEntry
	// @param header
	//
	OnOutgoingNack(egress *lf.LogicFace, pitEntry *table.PITEntry, header *component.NackHeader)

	//
	// 处理一个 CPacket 到来 （Incoming CPacket Pipeline）
	//
	// @Description:
	// @param ingress
	// @param cPacket
	//
	OnIncomingCPacket(ingress *lf.LogicFace, cPacket *packet.CPacket)

	//
	// 处理一个 CPacket 发出 （Outgoing CPacket Pipeline）
	//
	// @Description:
	// @param egress
	// @param cPacket
	//
	OnOutgoingCPacket(egress *lf.LogicFace, cPacket *packet.CPacket)
}
