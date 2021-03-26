//
// @Author: weiguohua
// @Description:
// @Version: 1.0.0
// @Date: 2021/3/16 下午3:49
// @Copyright: MIN-Group；国家重大科技基础设施——未来网络北大实验室；深圳市信息论与未来网络重点实验室
//
package lf

import (
	"log"
	"net"
	"time"
)

//
// @Description:  保存网卡信息的结构体
//
type NetIfInfo struct {
	name        string           // 网卡名
	macAddr     net.HardwareAddr // MAC地址
	state       bool             // true 为开启、 false为关闭
	mtu         int
	logicFaceId uint64 // 对应的逻辑接口号
}

//
// @Description: 监听所有网卡信息，使用协程来从每个网卡读取包
//
type EthernetListener struct {
	mDevices map[string]NetIfInfo // 用于保存，已经打开了的网卡的信息，以及相应的logicFace号
}

//
// @Description: 初始化对象
// @receiver e
//
func (e *EthernetListener) Init() {
	e.mDevices = make(map[string]NetIfInfo)
}

//
// @Description:  启动所有的协程
// @receiver e
//
func (e *EthernetListener) Start() {
	go e.monitorDev()
}

//
// @Description: 从全局logicFaceTable中删除网卡对应的logicFace，从设备列表中删除网卡信息
// @receiver e
// @param netInfo	网卡信息结构体
//
func (e *EthernetListener) deleteEtherFace(netInfo *NetIfInfo) {
	logicFace := GLogicFaceTable.GetLogicFacePtrById(netInfo.logicFaceId)
	if logicFace != nil {
		logicFace.Shutdown()
	}
	GLogicFaceTable.RemoveByLogicFaceId(netInfo.logicFaceId)
	delete(e.mDevices, netInfo.name)
}

//
// @Description: 更新网卡信息，如果网卡不存在列表里，则往列表里添加网卡，并创建一个协程，添加一个Ether类型的LogicFace。
//					如果网卡存在列表里，但是当前状态是关闭的，则关闭协程，并从列表里删除网卡
// @receiver e
// @param name		网卡名称
// @param macAddr	mac地址
// @param mtu		网卡MTU
// @param flag		网卡状态信息
//
func (e *EthernetListener) updateDev(name string, macAddr net.HardwareAddr, mtu int, flag net.Flags) {
	netInfo, ok := e.mDevices[name]
	if ok {
		if netInfo.state && (flag&net.FlagUp) == 0 {
			e.deleteEtherFace(&netInfo)
		}
		return
	}
	if (flag & net.FlagUp) != 0 {
		e.CreateEtherLogicFace(name, macAddr, mtu) // 创建以太网类型的 LogicFace
	}
}

//
// @Description: 	创建一个以太网类型的Face
// @receiver e
// @param ifName	网卡名
// @param macAddr	网卡Mac地址
// @param mtu		网卡MTU
// @return uint64	返回分配的logicFaceId
//
func (e *EthernetListener) CreateEtherLogicFace(ifName string, macAddr net.HardwareAddr, mtu int) uint64 {
	var netIfInfo NetIfInfo
	netIfInfo.name = ifName
	netIfInfo.macAddr = macAddr
	netIfInfo.state = true
	netIfInfo.mtu = mtu
	remoteMacAddr, _ := net.ParseMAC("01:00:5e:00:17:aa")
	logicFacePtr, logicFaceId := createEtherLogicFace(ifName, macAddr, remoteMacAddr, mtu)
	logicFacePtr.Start()
	netIfInfo.logicFaceId = logicFaceId
	e.mDevices[ifName] = netIfInfo
	return logicFaceId
}

//
// @Description: 	每2秒扫描一次主机网卡状态
// @receiver e
//
func (e *EthernetListener) monitorDev() {
	for true {
		interfaces, err := net.Interfaces()
		if err != nil {
			log.Fatal(err)
		}
		for _, d := range interfaces {
			e.updateDev(d.Name, d.HardwareAddr, d.MTU, d.Flags)
		}
		time.Sleep(time.Duration(2) * time.Second)
	}
}
