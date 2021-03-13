//
// @Author: Jianming Que
// @Description:
// @Version: 1.0.0
// @Date: 2021/3/3 3:21 下午
// @Copyright: MIN-Group；国家重大科技基础设施——未来网络北大实验室；深圳市信息论与未来网络重点实验室
//
package fw

import (
	"minlib/component"
	"minlib/packet"
	"mir-go/daemon/lf"
	"mir-go/daemon/table"
)

type IStrategy interface {
	//////////////////////////////////////////////////////////////////////////////////////////////////////
	//// Triggers
	//////////////////////////////////////////////////////////////////////////////////////////////////////

	//
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
	AfterReceiveInterest(ingress *lf.LogicFace, interest *packet.Interest, pitEntry *table.PITEntry)

	//
	// 当兴趣包命中缓存时，会触发本触发器
	//
	// @Description:
	//
	// @param ingress		Interest到来的入口LogicFace
	// @param data			缓存中得到的可以满足 Interest 的 Data
	// @param entry			兴趣包对应的PIT条目
	//
	AfterContentStoreHit(ingress *lf.LogicFace, data *packet.Data, entry *table.PITEntry)

	//
	// 当收到一个 Data 时，会触发本触发器
	//
	// @Description:
	//	Data 应当满足下列条件：
	//		- Data 被验证过可以匹配对应的PIT条目
	//		- Data 位于当前策略的命名空间下
	// @param ingress		Data 到来的入口 LogicFace
	// @param data			收到的 Data
	// @param pitEntry		Data 对应匹配的PIT条目
	//
	AfterReceiveData(ingress *lf.LogicFace, data *packet.Data, pitEntry *table.PITEntry)

	//
	// 当收到一个 Nack 时，会触发本触发器
	//
	// @Description:
	//
	// @param ingress		Nack 到来的入口 LogicFace
	// @param nack			收到的 Nack
	// @param pitEntry		Nack 对应匹配的PIT条目
	//
	AfterReceiveNack(ingress *lf.LogicFace, nack *packet.Nack, pitEntry *table.PITEntry)

	//
	// 当收到一个 CPacket 时，会触发本触发器
	//
	// @Description:
	// @param ingress		CPacket 到来的入口 LogicFace
	// @param cPacket		收到的 CPacket
	//
	AfterReceiveCPacket(ingress *lf.LogicFace, cPacket *packet.CPacket)

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
	sendInterest(egress *lf.LogicFace, interest *packet.Interest, pitEntry *table.PITEntry)

	//
	// 将 Data 从指定的逻辑接口转发出去
	//
	// @Description:
	// @param egress		转发 Data 的出口 LogicFace
	// @param data			要转发的 Data
	// @param pitEntry		Data 对应匹配的 PIT 条目
	//
	sendData(egress *lf.LogicFace, data *packet.Data, pitEntry *table.PITEntry)

	//
	// 将 Data 发送给对应 PIT 条目记录的所有符合条件的下游节点
	//
	// @Description:
	// @param ingress		Data 到来的入口 LogicFace => 主要是用来避免往收到 Data 包的 LogicFace 转发 Data
	// @param data			要转发的 Data
	// @param pitEntry		Data 对应匹配的 PIT 条目
	//
	sendDataToAll(ingress *lf.LogicFace, data *packet.Data, pitEntry *table.PITEntry)

	//
	// 往指定的逻辑接口发送一个 Nack
	//
	// @Description:
	// @param egress		转发 Nack 的出口 LogicFace
	// @param nackHeader	要转发出的Nack的元信息
	// @param pitEntry		Nack 对应匹配的 PIT 条目
	//
	sendNack(egress *lf.LogicFace, nackHeader *component.NackHeader, pitEntry *table.PITEntry)

	//
	// 将 Nack 发送给对应 PIT 条目记录的所有符合条件的下游节点
	//
	// @Description:
	// @param ingress		收到 Nack 的入口 LogicFace
	// @param nackHeader	要转发出的Nack的元信息
	// @param pitEntry		Nack 对应匹配的 PIT 条目
	//
	sendNackToAll(ingress *lf.LogicFace, nackHeader *component.NackHeader, pitEntry *table.PITEntry)

	//
	// 往指定的逻辑接口发送一个 CPacket
	//
	// @Description:
	// @param egress		转发 CPacket 的出口 LogicFace
	// @param cPacket		要转发出的 CPacket
	//
	sendCPacket(egress *lf.LogicFace, cPacket *packet.CPacket)

	//////////////////////////////////////////////////////////////////////////////////////////////////////
	//// 其它辅助函数
	//////////////////////////////////////////////////////////////////////////////////////////////////////

	//
	// 在 FIB 表中查询可用于转发 Interest 的 FIB 条目
	//
	// @Description:
	// @param interest
	//
	lookupFibForInterest(interest *packet.Interest)

	//
	// 在 FIB 表中查询可用于转发 CPacket 的 FIB 条目
	//
	// @Description:
	// @param cPacket
	//
	lookupFibForCPacket(cPacket *packet.CPacket)
}
