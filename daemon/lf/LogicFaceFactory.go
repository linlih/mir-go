// Package lf
// @Author: weiguohua
// @Description:
// @Version: 1.0.0
// @Date: 2021/3/17 上午10:33
// @Copyright: MIN-Group；国家重大科技基础设施——未来网络北大实验室；深圳市信息论与未来网络重点实验室
//
package lf

import (
	"github.com/google/gopacket/pcap"
	"minlib/logicface"
	"minlib/packet"
	"net"
)

//
// @Description: 创建一个以太网类型的LogicFace，并将创建的logicFace加入logicFace表中
// @param ifName	网卡名
// @param localMacAddr		网卡Mac地址
// @param remoteMacAddr		对端Mac地址
// @param mtu				网卡Mtu
// @return *LogicFace 		LogicFace指针
// @return *pcap.Handle		    pcap IO 句柄
//
func createEtherLogicFace(ifName string, localMacAddr, remoteMacAddr net.HardwareAddr, mtu int) (*LogicFace, *pcap.Handle) {
	var etherTransport EthernetTransport
	var logicFace LogicFace
	var linkService LinkService
	etherTransport.Init(ifName, localMacAddr, remoteMacAddr)
	if etherTransport.status == false {
		return nil, nil
	}
	linkService.Init(mtu)
	linkService.transport = &etherTransport
	linkService.logicFace = &logicFace
	etherTransport.linkService = &linkService
	logicFace.Init(&etherTransport, &linkService, LogicFaceTypeEther)
	gLogicFaceSystem.logicFaceTable.AddLogicFace(&logicFace)

	return &logicFace, etherTransport.handle
}

//
// @Description: 创建一个TCP类型的LogicFace
// @param conn	TCP连接句柄
// @return *LogicFace	LogicFace指针
// @return uint64		    LogicFace ID号
//
func createTcpLogicFace(conn net.Conn) (*LogicFace, uint64) {
	var tcpTransport TcpTransport
	var linkService LinkService
	var logicFace LogicFace

	tcpTransport.Init(conn)
	linkService.Init(9000)

	linkService.transport = &tcpTransport
	linkService.logicFace = &logicFace

	tcpTransport.linkService = &linkService

	logicFace.Init(&tcpTransport, &linkService, LogicFaceTypeTCP)
	logicFaceId := gLogicFaceSystem.logicFaceTable.AddLogicFace(&logicFace)
	return &logicFace, logicFaceId
}

//
// @Description: 创建一个unix socket类型的LogicFace
// @param conn	unix socket 连接句柄
// @return *LogicFace	LogicFace指针
// @return uint64		    LogicFace ID号
//
func createUnixLogicFace(conn net.Conn) (*LogicFace, uint64) {
	var unixTransport UnixStreamTransport
	var linkService LinkService
	var logicFace LogicFace

	unixTransport.Init(conn)
	linkService.Init(9000)

	linkService.transport = &unixTransport
	linkService.logicFace = &logicFace

	unixTransport.linkService = &linkService

	logicFace.Init(&unixTransport, &linkService, LogicFaceTypeUnix)
	logicFaceId := gLogicFaceSystem.logicFaceTable.AddLogicFace(&logicFace)
	return &logicFace, logicFaceId
}

//
// @Description: 创建一个Udp类型的LogicFace，UDP类型的logicFace都是只能用来发包
// @param conn	Udp句柄
// @param remoteAddr	对端udp地址
// @return *LogicFace	LogicFace 指针
// @return uint64		LogicFace ID号
//
func createUdpLogicFace(conn *net.UDPConn, remoteAddr *net.UDPAddr) (*LogicFace, uint64) {
	var udpTransport UdpTransport
	var linkService LinkService
	var logicFace LogicFace

	udpTransport.Init(conn, remoteAddr)
	linkService.Init(9000)

	linkService.transport = &udpTransport
	linkService.logicFace = &logicFace

	udpTransport.linkService = &linkService

	logicFace.Init(&udpTransport, &linkService, LogicFaceTypeUDP)
	logicFaceId := gLogicFaceSystem.logicFaceTable.AddLogicFace(&logicFace)
	return &logicFace, logicFaceId
}

//
// @Description: 创建一对相互收发包的内部logicFace，　需要调用者自己把要收包的logicface start 起来
// @return *LogicFace	 转发器使用的logicFace
// @return *logicface.LogicFace	其它模使用的logicFace
// @return *
//
func createInnerLogicFacePair() (*LogicFace, *logicface.LogicFace) {
	chan1 := make(chan *packet.LpPacket)
	chan2 := make(chan *packet.LpPacket)
	var innerTransport InnerTransport
	var linkService LinkService
	var newLogicFace LogicFace
	innerTransport.Init(chan1, chan2) // chan1 用于发包，　chan2用于收包
	linkService.Init(9000)
	linkService.transport = &innerTransport
	linkService.logicFace = &newLogicFace
	innerTransport.linkService = &linkService
	newLogicFace.Init(&innerTransport, &linkService, LogicFaceTypeInner)
	gLogicFaceSystem.logicFaceTable.AddLogicFace(&newLogicFace)

	var clientLogicFace logicface.LogicFace
	_ = clientLogicFace.InitWithInnerChan(chan2, chan1)
	newLogicFace.SetPersistence(1)

	return &newLogicFace, &clientLogicFace
}
