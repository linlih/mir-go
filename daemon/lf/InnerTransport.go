package lf

import (
	"minlib/packet"
	"mir-go/daemon/common"
)

//
// @Description:  用于MIR内部模块间通信的，使用 chan 进行通信， 使用一个用于发数据的chan和
type InnerTransport struct {
	Transport
	sendChan chan<- *packet.LpPacket
	recvChan <-chan *packet.LpPacket
}

func (i *InnerTransport) Init(sendChan chan<- *packet.LpPacket, recvChan <-chan *packet.LpPacket) {
	i.sendChan = sendChan
	i.recvChan = recvChan
	i.localAddr = "nil"
	i.localUri = "nil"
	i.remoteAddr = "inner://nil"
	i.remoteAddr = "inner://nil"
}

//
// @Description: 只能关闭发数据的chan
// @receiver i
//
func (i *InnerTransport) Close() {
	close(i.sendChan)
}

//
// @Description: 往chan里发lpPacket
// @receiver i
// @param lpPacket
//
func (i *InnerTransport) Send(lpPacket *packet.LpPacket) {
	i.sendChan <- lpPacket
}

//
// @Description: 从chan中接收lpPacket
// @receiver i
//
func (i *InnerTransport) Receive() {
	for true {
		lpPacket, ok := <-i.recvChan
		if !ok {
			common.LogInfo("inner channel has been closed")
			break
		}
		i.linkService.ReceivePacket(lpPacket)
	}
}
