//
// @Author: weiguohua
// @Description:
// @Version: 1.0.0
// @Date: 2021/3/16 上午11:35
// @Copyright: MIN-Group；国家重大科技基础设施——未来网络北大实验室；深圳市信息论与未来网络重点实验室
//
package lf

import (
	"minlib/packet"
	"mir-go/daemon/common"
	"net"
)

//
// @Description:  Udp通信隧道
//
type UdpTransport struct {
	Transport
	conn          *net.UDPConn // UDP句柄
	remoteUdpAddr net.UDPAddr  // 对端UDP地址，用于发送UDP包
}

//
// @Description: 	用于接收数据的transport的初始化函数
// @receiver u
// @param conn		UDP句俩
// @param remoteUdpAddr		对端udp地址
//
func (u *UdpTransport) Init(conn *net.UDPConn, remoteUdpAddr *net.UDPAddr) {
	u.conn = conn
	u.localAddr = conn.LocalAddr().String()
	u.localUri = "udp://" + u.localAddr
	if remoteUdpAddr == nil {
		u.remoteAddr = "nil"
		u.remoteUri = "udp://nil"
	} else {
		u.remoteAddr = remoteUdpAddr.String()
		u.remoteUri = "udp://" + u.remoteAddr
		u.remoteUdpAddr = *remoteUdpAddr
	}
}

//
// @Description: 关闭函数
// @receiver u
//
func (u *UdpTransport) Close() {
	err := u.conn.Close()
	if err != nil {
		common.LogWarn(err)
	}
}

//
// @Description: 往remoteUdpAddr这个地址发送一个UDP包，UDP包里装着一个lpPacket
// @receiver u
// @param lpPacket
//
func (u *UdpTransport) Send(lpPacket *packet.LpPacket) {
	encodeBufLen, encodeBuf := encodeLpPacket2ByteArray(lpPacket)
	if encodeBufLen <= 0 {
		return
	}
	_, err := u.conn.WriteToUDP(encodeBuf, &u.remoteUdpAddr)
	if err != nil {
		common.LogWarn(err)
	}
}

//
// @Description: 从UDP句柄中接收到UDP包，并处理
// @receiver u
//
func (u *UdpTransport) doReceive() {
	// 目前用不到
}

//
// @Description: 通过创建一个协程来负责调用这个函数
// @receiver u
//
func (u *UdpTransport) Receive() {
	// TODO 目前用不到
}
