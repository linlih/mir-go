package fw

import (
	"minlib/component"
	"minlib/packet"
	"mir/daemon/lf"
	"mir/daemon/table"
)

//
// MIR 转发器实例
//
// @Description:
//
type Forwarder struct {
}

func (f *Forwarder) OnIncomingInterest(ingress *lf.LogicFace, interest *packet.Interest) {
	panic("implement me")
}

func (f *Forwarder) OnInterestLoop(ingress *lf.LogicFace, interest *packet.Interest) {
	panic("implement me")
}

func (f *Forwarder) OnContentStoreMiss(ingress *lf.LogicFace, pitEntry *table.PITEntry, interest *packet.Interest) {
	panic("implement me")
}

func (f *Forwarder) OnContentStoreHit(ingress *lf.LogicFace, pitEntry *table.PITEntry, interest *packet.Interest, data *packet.Data) {
	panic("implement me")
}

func (f *Forwarder) OnOutgoingInterest(egress *lf.LogicFace, pitEntry *table.PITEntry, interest *packet.Interest) {
	panic("implement me")
}

func (f *Forwarder) OnInterestFinalize(pitEntry *table.PITEntry) {
	panic("implement me")
}

func (f *Forwarder) OnIncomingData(ingress *lf.LogicFace, data *packet.Data) {
	panic("implement me")
}

func (f *Forwarder) OnDataUnsolicited(ingress *lf.LogicFace, data *packet.Data) {
	panic("implement me")
}

func (f *Forwarder) OnOutgoingData(egress *lf.LogicFace, data *packet.Data) {
	panic("implement me")
}

func (f *Forwarder) OnIncomingNack(ingress *lf.LogicFace, nack *packet.Nack) {
	panic("implement me")
}

func (f *Forwarder) OnOutgoingNack(egress *lf.LogicFace, pitEntry *table.PITEntry, header *component.NackHeader) {
	panic("implement me")
}

func (f *Forwarder) OnIncomingCPacket(ingress *lf.LogicFace, cPacket *packet.CPacket) {
	panic("implement me")
}

func (f *Forwarder) OnOutgoingCPacket(egress *lf.LogicFace, cPacket *packet.CPacket) {
	panic("implement me")
}
