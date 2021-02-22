package fw

import "mir/daemon/common"

//
// MIR 转发器实例
//
// @Description:
//
type Forwarder struct {
}

func (f *Forwarder) IncomingInterest() {

}

func OnIncomingInterest() {
	common.LogFatal("233")
}
