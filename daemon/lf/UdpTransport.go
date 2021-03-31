//
// @Author: weiguohua
// @Description:
// @Version: 1.0.0
// @Date: 2021/3/16 上午11:35
// @Copyright: MIN-Group；国家重大科技基础设施——未来网络北大实验室；深圳市信息论与未来网络重点实验室
//
package lf

import (
	"log"
	"minlib/packet"
	"net"
)

//
// @Description:  Udp通信隧道
//
type UdpTransport struct {
	Transport
	conn            *net.UDPConn           // UDP句柄
	recvBuf         []byte                 // 接收缓冲区，大小为  9000
	remoteUdpAddr   net.UDPAddr            // 对端UDP地址，用于发送UDP包
	udpTransportMap *map[string]*LogicFace // remoteUri 和 logicFace的映射表
}

//
// @Description: 	用于接收数据的transport的初始化函数
// @receiver u
// @param conn		UDP句俩
// @param udpTransportMap		remoteUri 和 logicFace的映射表
//
func (u *UdpTransport) Init(conn *net.UDPConn, udpTransportMap *map[string]*LogicFace) {
	u.conn = conn
	u.localAddr = conn.LocalAddr().String()
	u.localUri = "udp://" + u.localAddr
	if conn.RemoteAddr() == nil {
		u.remoteAddr = "nil"
		u.remoteUri = "udp://" + u.remoteAddr
	} else {
		u.remoteAddr = conn.RemoteAddr().String()
		u.remoteUri = "udp://" + u.remoteAddr
	}
	u.recvBuf = make([]byte, 9000)
	u.udpTransportMap = udpTransportMap
}

//
// @Description: 用于发送数据的udp transport 初始化函数
// @receiver u
// @param conn	UDP句俩
// @param remoteUdpAddr		对端的udp地址
//
func (u *UdpTransport) InitHalf(conn *net.UDPConn, remoteUdpAddr *net.UDPAddr) {
	u.conn = conn
	u.localUri = "udp://" + conn.LocalAddr().String()
	u.remoteUri = "udp://" + remoteUdpAddr.String()
	u.remoteUdpAddr = *remoteUdpAddr
}

//
// @Description: 关闭函数
// @receiver u
//
func (u *UdpTransport) Close() {
	err := u.conn.Close()
	if err != nil {
		log.Println(err)
	}
}

//
// @Description: 往remoteUdpAddr这个地址发送一个UDP包，UDP包里装着一个lpPacket
// @receiver u
// @param lpPacket
//
func (u *UdpTransport) Send(lpPacket *packet.LpPacket) {
	encodeBufLen, encodeBuf := u.encodeLpPacket2ByteArray(lpPacket)
	if encodeBufLen <= 0 {
		return
	}
	_, err := u.conn.WriteToUDP(encodeBuf, &u.remoteUdpAddr)
	if err != nil {
		log.Println(err)
	}
}

//
// @Description:  这里与EtherTransport不同，
//			UdpTransport如果收到一个UDP包，而这个包的源UDP地址在udpTransportMap不存在，则Transport会创建一个新的LogicFace，并以这个
//			UDP包的源地址作为key，将新创建的logicFace放到udpTransportMap中。
//			TODO 这里可能需要改进，改成由上层的管理模块验证UDP包的可靠性之后，再创建UDP类型的LogicFace
// @receiver u
// @param remoteAddr
// @return *LinkService
//
func (u *UdpTransport) GetOrCreateLinkService(remoteAddr *net.UDPAddr) *LinkService {
	logicFace, ok := (*u.udpTransportMap)[remoteAddr.String()]
	if ok {
		return logicFace.linkService
	}
	logicFace, _ = createHalfUdpLogicFace(u.conn, remoteAddr)
	return logicFace.linkService
}

//
// @Description: 从UDP句柄中接收到UDP包，并处理
// @receiver u
//
func (u *UdpTransport) doReceive() {
	readLen, remoteUdpAddr, err := u.conn.ReadFromUDP(u.recvBuf)
	if err != nil {
		log.Println(err)
		return
	}
	log.Println("recv from : ", remoteUdpAddr)
	lpPacket, err := u.parseByteArray2LpPacket(u.recvBuf[:readLen])
	if err != nil {
		log.Println(err)
		return
	}
	// TODO 先验证再添加LogicFace
	linkService := u.GetOrCreateLinkService(remoteUdpAddr)
	linkService.ReceivePacket(lpPacket)
}

//
// @Description: 通过创建一个协程来负责调用这个函数
// @receiver u
//
func (u *UdpTransport) Receive() {
	for true {
		u.doReceive()
	}
}
