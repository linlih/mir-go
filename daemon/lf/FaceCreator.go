//
// @Author: weiguohua
// @Description:
// @Version: 1.0.0
// @Date: 2021/3/23 上午10:46
// @Copyright: MIN-Group；国家重大科技基础设施——未来网络北大实验室；深圳市信息论与未来网络重点实验室
//
package lf

import (
	"errors"
	"log"
	"net"
)

//
// @Description: 外部接口，为其他模块提供创建各种接口的函数
//

//
// @Description:	给其他模块调用，创建一个Ether类型的LogicFace，传入本地网卡地址和目的MAC地址。
//					本函数会：
//					（1）判断网卡是否存在，网卡是否已经启动； 如果网卡没启动，则返回错误信息；
//					（2）判断 该本地网卡和目的MAC 对应的logicFace 是否已经存在，如果已经存在，则返回已经存在的logicFaceId；
//					（3） 如果不存在，调用内部接口， 创建一个只用于发送数据的LogicFace。
//					注意 ： 本函数不会判断传输的对端MAC地址是否真实存在。
// @param localIfName	本地网卡名
// @param remoteMacAddr		对端MAC地址
// @return uint64		logicFaceId
// @return error		错误信息
//
func CreateEtherLogicFace(localIfName string, remoteMacAddr net.HardwareAddr) (uint64, error) {

	netIfInfo, ok := gLogicFaceSystem.ethernetListener.mDevices[localIfName]
	if !ok {
		return 0, errors.New("can not find local dev name : " + localIfName)
	}

	logicFace, ok := (*gEtherAddrFaceMap)[netIfInfo.macAddr.String()+"-"+remoteMacAddr.String()]
	if ok {
		return logicFace.LogicFaceId, nil
	}

	logicFace, logicFaceId := createEtherLogicFace(localIfName, netIfInfo.macAddr, remoteMacAddr, netIfInfo.mtu)

	//key 的格式是收到以太网帧的 "<目的MAC地址>-<源MAC地址>"
	(*gEtherAddrFaceMap)[netIfInfo.macAddr.String()+"-"+remoteMacAddr.String()] = logicFace

	return logicFaceId, nil
}

//
// @Description:  给其他模块调用，创建一个TCP类型的LogicFace，传入对方的TCP地址，格式是 "<ip>:<port>"，如"192.168.3.7:13899"。
//				函数会执行以下操作：
//				（1） 尝试连接远程TCP地址，如果连接不成功，则返回连接错误信息
//				（2） 如果连接成功，调用内部函数，创建一个TCP类型的logicFace
//				（3） 启动该logicFace的接收数据协程
// @param remoteUri		对方的TCP地址，格式是 "<ip>:<port>"，如"192.168.3.7:13899"
// @return uint64		logicFaceId
// @return error		错误信息
//
func CreateTcpLogicFace(remoteUri string) (uint64, error) {
	conn, err := net.Dial("tcp", remoteUri)
	if err != nil {
		log.Println(err)
		return 0, err
	}
	logicFace, logicFaceId := createTcpLogicFace(conn)
	logicFace.Start()
	return logicFaceId, nil
}

//
// @Description:	给其他模块调用，创建一个UDP类型的LogicFace，传入对方的UDP地址，格式是 "<ip>:<port>"，如"192.168.3.7:13899"
//				函数会执行以下操作：
//				（1） 尝试解析UDP地址，如果解析不成功，则返回连接错误信息
//				（2） 如果解析UDP地址成功，调用内部函数，创建一个UDP类型的logicFace
//				（3） 启动该logicFace的接收数据协程
// @param remoteUri
// @return uint64
// @return error
//
func CreateUdpLogicFace(remoteUri string) (uint64, error) {
	udpAddr, err := net.ResolveUDPAddr("udp4", remoteUri)
	if err != nil {
		log.Println(err)
		return 0, err
	}
	//udpConn, err := net.DialUDP("udp", nil, udpAddr)
	_, logicFaceId := createHalfUdpLogicFace(gLogicFaceSystem.udpListener.conn, udpAddr)
	return logicFaceId, nil
}

//
// @Description:  给其他模块调用，创建一个unix socket类型的LogicFace，传入对方的unix地址，格式是 文件路径，如"/tmp/mirsock"。
//				函数会执行以下操作：
//				（1） 尝试解析unix地址，如果解析不成功，则返回连接错误信息
//				（2） 尝试连接远程地址，如果连接不成功，则返回连接错误信息
//				（3） 如果连接成功，调用内部函数，创建一个unix类型的logicFace
//				（4） 启动该logicFace的接收数据协程
// @param remoteUri		传入对方的unix地址，格式是 文件路径，如"/tmp/mirsock"。
// @return uint64		logicFaceId
// @return error		错误信息
//
func CreateUnixLogicFace(remoteUri string) (uint64, error) {
	addr, err := net.ResolveUnixAddr("unix", remoteUri)
	if err != nil {
		log.Println(err)
		return 0, err
	}
	conn, err := net.DialUnix("unix", nil, addr)
	if err != nil {
		panic("DialUnix failed.")
	}
	logicFace, logicFaceId := createUnixLogicFace(conn)
	logicFace.Start()
	return logicFaceId, nil
}
