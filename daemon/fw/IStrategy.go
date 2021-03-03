//
// @Author: Jianming Que
// @Description:
// @Version: 1.0.0
// @Date: 2021/3/3 3:21 下午
// @Copyright: MIN-Group；国家重大科技基础设施——未来网络北大实验室；深圳市信息论与未来网络重点实验室
//
package fw

import (
	"minlib/packet"
	"mir/daemon/lf"
	"mir/daemon/table"
)

type IStrategy interface {
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
}
