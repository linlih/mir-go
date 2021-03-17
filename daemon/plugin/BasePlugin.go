//
// @Author: Jianming Que
// @Description:
// @Version: 1.0.0
// @Date: 2021/3/17 3:15 下午
// @Copyright: MIN-Group；国家重大科技基础设施——未来网络北大实验室；深圳市信息论与未来网络重点实验室
//
package plugin

import (
	"minlib/component"
	"minlib/packet"
	"mir-go/daemon/lf"
	"mir-go/daemon/table"
)

//
// 插件的基准实现，所有的锚点都返回默认的 0
//
// @Description:
//
type BasePlugin struct {
}

func (b BasePlugin) OnIncomingInterest(ingress *lf.LogicFace, interest *packet.Interest) int {
	return 0
}

func (b BasePlugin) OnInterestLoop(ingress *lf.LogicFace, interest *packet.Interest) int {
	return 0
}

func (b BasePlugin) OnContentStoreMiss(ingress *lf.LogicFace, pitEntry *table.PITEntry, interest *packet.Interest) int {
	return 0
}

func (b BasePlugin) OnContentStoreHit(ingress *lf.LogicFace, pitEntry *table.PITEntry, interest *packet.Interest, data *table.CSEntry) int {
	return 0
}

func (b BasePlugin) OnOutgoingInterest(egress *lf.LogicFace, pitEntry *table.PITEntry, interest *packet.Interest) int {
	return 0
}

func (b BasePlugin) OnInterestFinalize(pitEntry *table.PITEntry) int {
	return 0
}

func (b BasePlugin) OnIncomingData(ingress *lf.LogicFace, data *packet.Data) int {
	return 0
}

func (b BasePlugin) OnDataUnsolicited(ingress *lf.LogicFace, data *packet.Data) int {
	return 0
}

func (b BasePlugin) OnOutgoingData(egress *lf.LogicFace, data *packet.Data) int {
	return 0
}

func (b BasePlugin) OnIncomingNack(ingress *lf.LogicFace, nack *packet.Nack) int {
	return 0
}

func (b BasePlugin) OnOutgoingNack(egress *lf.LogicFace, pitEntry *table.PITEntry, header *component.NackHeader) int {
	return 0
}

func (b BasePlugin) OnIncomingCPacket(ingress *lf.LogicFace, cPacket *packet.CPacket) int {
	return 0
}

func (b BasePlugin) OnOutgoingCPacket(egress *lf.LogicFace, cPacket *packet.CPacket) int {
	return 0
}