//
// @Author: weiguohua
// @Description:
// @Version: 1.0.0
// @Date: 2021/3/16 下午3:49
// @Copyright: MIN-Group；国家重大科技基础设施——未来网络北大实验室；深圳市信息论与未来网络重点实验室
//
package lf

import (
	"mir-go/daemon/common"
	"net"
	"time"
)

//
// @Description: 监听所有网卡信息，使用协程来从每个网卡读取包，每个网卡对应一个 InterfaceListener， 一个 InterfaceListener
//			用于专门接收一个网卡的包
//
type EthernetListener struct {
	mInterfaceListeners map[string]*InterfaceListener // 用于保存，已经打开了的网卡的信息，以及相应的logicFace号
	badDev              map[string]int                // 用于保存无法启动的网卡名
}

//
// @Description: 初始化对象
// @receiver e
//
func (e *EthernetListener) Init() {
	e.mInterfaceListeners = make(map[string]*InterfaceListener)
	e.badDev = make(map[string]int)
}

//
// @Description:  启动所有的协程
// @receiver e
//
func (e *EthernetListener) Start() {
	go e.monitorDev()
}

//
// @Description:  关闭网卡监听器
//			从全局logicFaceTable中删除网卡对应的logicFace，从设备列表中删除网卡信息
// @receiver e
// @param netInfo	网卡信息结构体
//
func (e *EthernetListener) closeInterfaceListener(ifListener *InterfaceListener) {
	ifListener.Close() // 关闭这个网卡对应的所有logicFace
	delete(e.mInterfaceListeners, ifListener.name)
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
	ifListener, ok := e.mInterfaceListeners[name]
	if ok {
		if ifListener.state && (flag&net.FlagUp) == 0 {
			e.closeInterfaceListener(ifListener)
		}
		return
	}
	if (flag & net.FlagUp) != 0 {
		_, ok = e.badDev[name]
		if ok { // 该网卡前面尝试启动过，启动不了
			return
		}
		e.CreateInterfaceListener(name, macAddr, mtu) // 创建以太网类型的 LogicFace
	}
}

//
// @Description: 	创建一个网卡监听器，用于从特定的网卡读取网络包
// @receiver e
// @param ifName	网卡名
// @param macAddr	网卡Mac地址
// @param mtu		网卡MTU
//
func (e *EthernetListener) CreateInterfaceListener(ifName string, macAddr net.HardwareAddr, mtu int) {
	var ifListener InterfaceListener
	ifListener.Init(ifName, macAddr, mtu)
	err := ifListener.Start() // 启动从网卡读取包的协程
	if err != nil {
		e.badDev[ifName] = 1
		return
	}
	e.mInterfaceListeners[ifName] = &ifListener
}

//
// @Description: 	每2秒扫描一次主机网卡状态
// @receiver e
//
func (e *EthernetListener) monitorDev() {
	for true {
		interfaces, err := net.Interfaces()
		if err != nil {
			common.LogFatal(err)
		}
		for _, d := range interfaces {
			e.updateDev(d.Name, d.HardwareAddr, d.MTU, d.Flags)
		}
		time.Sleep(time.Duration(2) * time.Second)
	}
}

//
// @Description: 	删除一个logicFace
// @receiver e
// @param localMacAddr	本地mac地址
// @param remoteMacAddr	对端mac地址
//
func (e *EthernetListener) DeleteLogicFace(localMacAddr string, remoteMacAddr string) {
	for _, ifListener := range e.mInterfaceListeners {
		if ifListener.macAddr.String() == localMacAddr {
			ifListener.DeleteLogicFace(remoteMacAddr)
			return
		}
	}
}
