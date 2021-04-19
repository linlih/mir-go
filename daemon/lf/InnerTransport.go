package lf

import (
	common2 "minlib/common"
	"minlib/packet"
)

// InnerTransport
// @Description:  用于MIR内部模块间通信的，使用 chan 进行通信， 使用一个用于发数据的chan和
type InnerTransport struct {
	Transport
	sendChan chan<- *packet.LpPacket
	recvChan <-chan *packet.LpPacket
}

func (i *InnerTransport) Init(sendChan chan<- *packet.LpPacket, recvChan <-chan *packet.LpPacket) {
	i.sendChan = sendChan
	i.recvChan = recvChan
	i.localAddr = "inner://nil"
	i.localUri = "inner://nil"
	i.remoteAddr = "inner://nil"
	i.remoteUri = "inner://nil"
}

// Close
// @Description: 只能关闭发数据的chan
// @receiver i
//
func (i *InnerTransport) Close() {
	close(i.sendChan)
}

// Send
// @Description: 往chan里发lpPacket
// @receiver i
// @param lpPacket
//
func (i *InnerTransport) Send(lpPacket *packet.LpPacket) {
	i.sendChan <- lpPacket
}

// Receive
// @Description: 从chan中接收lpPacket
// @receiver i
//
func (i *InnerTransport) Receive() {
	for true {
		lpPacket, ok := <-i.recvChan
		if !ok {
			common2.LogInfo("inner channel has been closed")
			break
		}
		i.linkService.ReceivePacket(lpPacket)
	}
}
