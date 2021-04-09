//
// @Author: weiguohua
// @Description:
// @Version: 1.0.0
// @Date: 2021/4/1 下午2:29
// @Copyright: MIN-Group；国家重大科技基础设施——未来网络北大实验室；深圳市信息论与未来网络重点实验室
//
package lf

import (
	"errors"
	"github.com/google/gopacket"
	"github.com/google/gopacket/pcap"
	"minlib/packet"
	"mir-go/daemon/common"
	"net"
)

//
// @Description:  etherFaceMap用于保存mac地址和LogicFace的映射关系表。
//			key 的格式是收到以太网帧的 "<源MAC地址>"
//			value 是logicFace对象指针
//		使用这个映射表的原因在于：
//		（1） 一个物理网卡可能会对应多个logicFace，在MIR启动的时候，我们会启动一个以 "01:00:5e:00:17:aa" 为目的MAC地址的LogicFace，
//			我们先将这个logicFace称为logicFace0,这个logicFace0用于
//			接收从该网卡收到的以太网帧，同时也可以使用这个LogicFace0向该物理网卡对应的以太网发送以太网帧。由于使用该logicFace0发送的以太网帧的
//			目的MAC地址是"01:00:5e:00:17:aa"，是一个组播地址，所以这个logicFace0发出的以太网帧会被物理网卡所在的以太网中的所有其他网卡接收。
//		（2） 有时候，我们会需要创建一个一对一的以太网类型的logicFace，这时候我们会新建一个logicFace对应一个物理网卡，为了方便说明，
//			我们将这个新建的logicFace称为logicFace1。logicFace1的目的MAC地址假设是"fc:aa:14:cf:a6:97"，这是一个确切的对应特定物理网卡的地址
//			由于我们已经启动了logicFace0来接收物理网卡收到的所有MIN网络包，包括本应发往logicFace1的包，所以我们在创建logicFace1时不再启动收包协程，
//			这个时候，如果logicFace0收到了本应发往logicFace1的网络包，logicFace0会需要通过查找gEtherAddrFaceMap这个映射表，知道要调用logicFace1
//			的收到函数来处理网络分组。
//		（3） 在创建logicFace1时，我们会为logicFace1的etherTransport新创建一个pcap的handle用于发送网络包。
//
type InterfaceListener struct {
	name         string           // 网卡名
	macAddr      net.HardwareAddr // MAC地址
	state        bool             // true 为开启、 false为关闭
	mtu          int
	logicFace    *LogicFace
	etherFaceMap map[string]*LogicFace // 对端mac地址和face对象映射表
	pcapHandle   *pcap.Handle          // pcap 抓包句柄
}

//
// @Description: 	初始化函数
// @receiver i
// @param name	网卡名
// @param macAddr	网卡mac地址
// @param mtu	网卡MTU
//
func (i *InterfaceListener) Init(name string, macAddr net.HardwareAddr, mtu int) {
	i.etherFaceMap = make(map[string]*LogicFace)
	i.name = name
	i.macAddr = macAddr
	i.mtu = mtu
	i.state = true
}

//
// @Description: 	关闭当前监听器，以及与该网卡相关的所有logicFace
// @receiver i
//
func (i *InterfaceListener) Close() {
	i.logicFace.Shutdown()
	for _, value := range i.etherFaceMap {
		value.Shutdown()
	}
}

// @Description: 启动当前监听器
// @receiver i
// @return error 启动失败则返回错误
//
func (i *InterfaceListener) Start() error {
	remoteMacAddr, _ := net.ParseMAC("01:00:5e:00:17:aa")
	var logicFacePtr *LogicFace
	logicFacePtr, i.pcapHandle = createEtherLogicFace(i.name, i.macAddr, remoteMacAddr, i.mtu)
	if logicFacePtr == nil {
		return errors.New("create ether logic face error")
	}
	i.logicFace = logicFacePtr
	go i.readPacketFromDev()
	return nil
}

//
// @Description: 	通过mac地址获得一个 LogicFace
// @receiver i
// @param macAddr
// @return *LogicFace
//
func (i *InterfaceListener) GetLogicFaceByMacAddr(macAddr string) *LogicFace {
	logicFace, ok := i.etherFaceMap[macAddr]
	if ok {
		return logicFace
	}
	return nil
}

//
// @Description: 	收到包的处理函数
// @receiver i
// @param lpPacket	收到的包
// @param srcMacAddr	收到包的源MAC地址
//
func (i *InterfaceListener) onReceive(lpPacket *packet.LpPacket, srcMacAddr string) {
	logicFace, ok := i.etherFaceMap[srcMacAddr]
	if ok {
		if logicFace.state == false {
			delete(i.etherFaceMap, srcMacAddr)
		}
		logicFace.linkService.ReceivePacket(lpPacket)
		return
	}
	remoteMacAddr, err := net.ParseMAC(srcMacAddr)
	if err != nil {
		common.LogWarn(err)
		return
	}
	if checkIdentity(lpPacket) == false {
		common.LogInfo("user identify verify no pass")
		return
	}
	logicFacePtr, _ := createEtherLogicFace(i.name, i.macAddr, remoteMacAddr, i.mtu)
	if logicFacePtr == nil {
		common.LogFatal("create ether logicface, error")
	}
	i.etherFaceMap[srcMacAddr] = logicFacePtr
	logicFacePtr.linkService.ReceivePacket(lpPacket)
}

//
// @Description: 	不断的从网卡中读包
// @receiver i
//
func (i *InterfaceListener) readPacketFromDev() {
	pktSrc := gopacket.NewPacketSource(i.pcapHandle, i.pcapHandle.LinkType())
	for pkt := range pktSrc.Packets() {
		lpPacket, err := parseByteArray2LpPacket(pkt.Data()[14:])
		if err != nil {
			common.LogError("parse byte to lpPacket error : ", err)
		} else {
			i.onReceive(lpPacket, pkt.LinkLayer().LinkFlow().Src().String())
		}
	}

}

//
// @Description: 通过mac地址删除一个etherFaceMap中的logicFace
// @receiver i
// @param remoteMacAddr
//
func (i *InterfaceListener) DeleteLogicFace(remoteMacAddr string) {
	delete(i.etherFaceMap, remoteMacAddr)
}

//
// @Description: 往etherFaceMap中添加一个 LogicFace
// @receiver i
// @param remoteMacAddr
// @param face
//
func (i *InterfaceListener) AddLogicFace(remoteMacAddr string, face *LogicFace) {
	i.etherFaceMap[remoteMacAddr] = face
}
