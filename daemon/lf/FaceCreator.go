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

func CreateEtherLogicFace(localIfName string, remoteMacAddr net.HardwareAddr) (uint64, error) {

	netIfInfo, ok := gLogicFaceSystem.ethernetListener.mDevices[localIfName]
	if !ok {
		return 0, errors.New("can not find local dev name : " + localIfName)
	}
	logicFace, logicFaceId := createEtherLogicFace(localIfName, netIfInfo.macAddr, remoteMacAddr, netIfInfo.mtu)
	(*gEtherAddrFaceMap)[remoteMacAddr.String()] = logicFace

	return logicFaceId, nil
}

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
