//
// @Author: Jianming Que
// @Description:
// @Version: 1.0.0
// @Date: 2021/3/4 10:18 上午
// @Copyright: MIN-Group；国家重大科技基础设施——未来网络北大实验室；深圳市信息论与未来网络重点实验室
//
package fw

import (
	"minlib/component"
	"minlib/packet"
	"mir/daemon/lf"
	"mir/daemon/table"
)

type Strategy struct {
}

func (s *Strategy) AfterReceiveInterest(ingress *lf.LogicFace, interest *packet.Interest, pitEntry *table.PITEntry) {
	panic("implement me")
}

func (s *Strategy) AfterContentStoreHit(ingress *lf.LogicFace, data *packet.Data, entry *table.PITEntry) {
	panic("implement me")
}

func (s *Strategy) AfterReceiveData(ingress *lf.LogicFace, data *packet.Data, pitEntry *table.PITEntry) {
	panic("implement me")
}

func (s *Strategy) AfterReceiveNack(ingress *lf.LogicFace, nack *packet.Nack, pitEntry *table.PITEntry) {
	panic("implement me")
}

func (s *Strategy) AfterReceiveCPacket(ingress *lf.LogicFace, cPacket *packet.CPacket) {
	panic("implement me")
}

func (s *Strategy) sendInterest(egress *lf.LogicFace, interest *packet.Interest, pitEntry *table.PITEntry) {
	panic("implement me")
}

func (s *Strategy) sendData(egress *lf.LogicFace, data *packet.Data, pitEntry *table.PITEntry) {
	panic("implement me")
}

func (s *Strategy) sendDataToAll(ingress *lf.LogicFace, data *packet.Data, pitEntry *table.PITEntry) {
	panic("implement me")
}

func (s *Strategy) sendNack(egress *lf.LogicFace, nackHeader *component.NackHeader, pitEntry *table.PITEntry) {
	panic("implement me")
}

func (s *Strategy) sendNackToAll(ingress *lf.LogicFace, nackHeader *component.NackHeader, pitEntry *table.PITEntry) {
	panic("implement me")
}

func (s *Strategy) sendCPacket(egress *lf.LogicFace, cPacket *packet.CPacket) {
	panic("implement me")
}

func (s *Strategy) lookupFibForInterest(interest *packet.Interest) {
	panic("implement me")
}

func (s *Strategy) lookupFibForCPacket(cPacket *packet.CPacket) {
	panic("implement me")
}
