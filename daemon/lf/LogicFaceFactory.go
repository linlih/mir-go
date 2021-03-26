//
// @Author: weiguohua
// @Description:
// @Version: 1.0.0
// @Date: 2021/3/17 上午10:33
// @Copyright: MIN-Group；国家重大科技基础设施——未来网络北大实验室；深圳市信息论与未来网络重点实验室
//
package lf

import "net"

//
// @Description: 创建一个以太网类型的LogicFace，并将创建的logicFace加入logicFace表中
// @param ifName	网卡名
// @param localMacAddr		网卡Mac地址
// @param remoteMacAddr		对端Mac地址
// @param mtu				网卡Mtu
// @return *LogicFace 		LogicFace指针
// @return uint64		    LogicFace ID号
//
func createEtherLogicFace(ifName string, localMacAddr, remoteMacAddr net.HardwareAddr, mtu int) (*LogicFace, uint64) {
	var etherTransport EthernetTransport
	var logicFace LogicFace
	var linkService LinkService

	//etherTransport.Init(ifName, localMacAddr, remoteMacAddr)
	linkService.Init(mtu)
	linkService.transport = &etherTransport
	linkService.logicFace = &logicFace
	etherTransport.linkService = &linkService
	logicFace.Init(&etherTransport, &linkService, LogicFaceTypeEther)
	logicFaceId := GLogicFaceTable.AddLogicFace(&logicFace)

	return &logicFace, logicFaceId
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
	logicFaceId := GLogicFaceTable.AddLogicFace(&logicFace)
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
	logicFaceId := GLogicFaceTable.AddLogicFace(&logicFace)
	return &logicFace, logicFaceId
}

//
// @Description: 创建一个Udp类型的LogicFace
// @param conn	Udp句柄
// @return *LogicFace	LogicFace 指针
// @return uint64		LogicFace ID号
//
func createUdpLogicFace(conn *net.UDPConn) (*LogicFace, uint64) {
	var udpTransport UdpTransport
	var linkService LinkService
	var logicFace LogicFace

	udpTransport.Init(conn, gUdpAddrFaceMap)
	linkService.Init(9000)

	linkService.transport = &udpTransport
	linkService.logicFace = &logicFace

	udpTransport.linkService = &linkService

	logicFace.Init(&udpTransport, &linkService, LogicFaceTypeUDP)
	logicFaceId := GLogicFaceTable.AddLogicFace(&logicFace)
	return &logicFace, logicFaceId
}

//
// @Description: 创建一个只用于发送数据的Udp类型的LogicFace
// @param conn	Udp句柄
// @return *LogicFace	LogicFace 指针
// @return uint64		LogicFace ID号
//
func createHalfUdpLogicFace(conn *net.UDPConn, remoteAddr *net.UDPAddr) (*LogicFace, uint64) {
	var udpTransport UdpTransport
	var linkService LinkService
	var logicFace LogicFace

	udpTransport.InitHalf(conn, remoteAddr)
	linkService.Init(9000)

	linkService.transport = &udpTransport
	linkService.logicFace = &logicFace

	udpTransport.linkService = &linkService

	logicFace.Init(&udpTransport, &linkService, LogicFaceTypeUDP)
	logicFaceId := GLogicFaceTable.AddLogicFace(&logicFace)
	(*gUdpAddrFaceMap)[remoteAddr.String()] = &logicFace
	return &logicFace, logicFaceId
}
