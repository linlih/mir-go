//
// @Author: weiguohua
// @Description:
// @Version: 1.0.0
// @Date: 2021/3/16 下午3:50
// @Copyright: MIN-Group；国家重大科技基础设施——未来网络北大实验室；深圳市信息论与未来网络重点实验室
//
package lf

import (
	"minlib/packet"
	"mir-go/daemon/common"
	"net"
	"strconv"
)

//
// @Description:  udpAddrFaceMap 用于保存IP：PORT信息和logicFace的映射关系
//			key 的格式是收到UDP包的 "<源IP地址>:<源端口号>"
//			value 的格式是logicFace对象指针
//		使用这个映射表的原因与gEtherAddrFaceMap类似
//		（1） 一个UDP端口13899可能会对应多个不同的logicFace。
//		（2）
//		（3） 在创建logicFace1时，与ether类型的logicFace不同的是，我们不会创建一个新的handle，而是一直使用logicFace0的handle。
//			TODO 这样做可能会有问题，现在还没考虑到，到时候改成新建一个handle也比较简单，现在先这么做
//
type UdpListener struct {
	udpPort        uint16
	conn           *net.UDPConn
	udpAddrFaceMap map[string]*LogicFace
	recvBuf        []byte // 接收缓冲区，大小为  9000
}

func (u *UdpListener) Init() {
	u.udpPort = 13899
	u.udpAddrFaceMap = make(map[string]*LogicFace)
	u.recvBuf = make([]byte, 9000)
}

//
// @Description: 创建一个udp类型的logicFace，并启动logicFace，启动一个协程负责接收UDP包
// @receiver t
// @param conn	新udp 句柄
//
func (u *UdpListener) createUdpLogicFace(conn *net.UDPConn) {
	createUdpLogicFace(conn, nil)
}

//
// @Description:  启动监听协程
// @receiver t
//
func (u *UdpListener) Start() {
	udpAddr, err := net.ResolveUDPAddr("udp4", "0.0.0.0:"+strconv.Itoa(int(u.udpPort)))
	if err != nil {
		common.LogFatal(err)
		return
	}
	conn, err := net.ListenUDP("udp", udpAddr)
	if err != nil {
		common.LogFatal(err)
	}
	u.conn = conn
	u.createUdpLogicFace(conn)
	go func() {
		for {
			u.doReceive()
		}
	}()
}

func (u *UdpListener) onReceive(lpPacket *packet.LpPacket, remoteUdpAddr *net.UDPAddr) {
	logicFace, ok := u.udpAddrFaceMap[remoteUdpAddr.String()]
	if ok {
		if logicFace.state == false {
			delete(u.udpAddrFaceMap, remoteUdpAddr.String())
			return
		}
		logicFace.linkService.ReceivePacket(lpPacket)
		return
	}
	// TODO 先验证再添加LogicFace
	logicFace, _ = createUdpLogicFace(u.conn, remoteUdpAddr)
	logicFace.linkService.ReceivePacket(lpPacket)
}

//
// @Description: 从UDP句柄中接收到UDP包，并处理
// @receiver u
//
func (u *UdpListener) doReceive() {
	readLen, remoteUdpAddr, err := u.conn.ReadFromUDP(u.recvBuf)
	if err != nil {
		common.LogWarn(err)
		return
	}
	common.LogDebug("recv from : ", remoteUdpAddr)
	lpPacket, err := parseByteArray2LpPacket(u.recvBuf[:readLen])
	if err != nil {
		common.LogWarn(err)
		return
	}
	u.onReceive(lpPacket, remoteUdpAddr)
}

func (u *UdpListener) DeleteLogicFace(remoteAddr string) {
	delete(u.udpAddrFaceMap, remoteAddr)
}

func (u *UdpListener) AddLogicFace(remoteAddr string, logicFace *LogicFace) {
	u.udpAddrFaceMap[remoteAddr] = logicFace
}
