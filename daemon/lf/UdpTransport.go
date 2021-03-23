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
	conn            *net.UDPConn
	recvBuf         []byte
	remoteUdpAddr   net.UDPAddr
	udpTransportMap *map[string]*LogicFace
}

func (u *UdpTransport) Init(conn *net.UDPConn, udpTransportMap *map[string]*LogicFace) {
	u.conn = conn
	u.localUri = "udp://" + conn.LocalAddr().String()
	u.remoteUri = "udp://" + conn.RemoteAddr().String()
	u.recvBuf = make([]byte, 9000)
	u.udpTransportMap = udpTransportMap
}

func (u *UdpTransport) InitHalf(conn *net.UDPConn, remoteUdpAddr *net.UDPAddr) {
	u.conn = conn
	u.localUri = "udp://" + conn.LocalAddr().String()
	u.remoteUri = "udp://" + remoteUdpAddr.String()
	u.remoteUdpAddr = *remoteUdpAddr
}

func (u *UdpTransport) Close() {
	err := u.conn.Close()
	if err != nil {
		log.Println(err)
	}
}

func (u *UdpTransport) Send(lpPacket *packet.LpPacket) {
	encodeBufLen, encodeBuf := u.encodeLpPacket2ByteArray(lpPacket)
	if encodeBufLen < 0 {
		return
	}
	_, err := u.conn.WriteToUDP(encodeBuf, &u.remoteUdpAddr)
	if err != nil {
		log.Println(err)
	}
}

func (u *UdpTransport) GetOrCreateLinkService(remoteAddr *net.UDPAddr) *LinkService {
	logicFace, ok := (*u.udpTransportMap)[remoteAddr.String()]
	if ok {
		return logicFace.linkService
	}
	logicFace, _ = createHalfUdpLogicFace(u.conn, remoteAddr)
	return logicFace.linkService
}

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

func (u *UdpTransport) Receive() {
	for true {
		u.doReceive()
	}
}

func (u *UdpTransport) GetRemoteUri() string {
	return u.remoteUri
}

func (u *UdpTransport) GetLocalUri() string {
	return u.localUri
}
