// Copyright [2022] [MIN-Group -- Peking University Shenzhen Graduate School Multi-Identifier Network Development Group]
//
// Licensed under the Apache License, Version 2.0 (the "License"): you may
// not use this file except in compliance with the License. You may obtain
// a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS, WITHOUT
// WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the
// License for the specific language governing permissions and limitations
// under the License.

// Package lf
// @Author: weiguohua
// @Description:
// @Version: 1.0.0
// @Date: 2021/4/1 下午2:29
// @Copyright: MIN-Group；国家重大科技基础设施——未来网络北大实验室；深圳市信息论与未来网络重点实验室
//
package lf

import (
	"encoding/binary"
	"errors"
	"github.com/google/gopacket"
	"github.com/google/gopacket/pcap"
	common2 "minlib/common"
	"minlib/packet"
	"mir-go/daemon/utils"
	"net"
)

// InterfaceListener
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
	name              string           // 网卡名
	macAddr           net.HardwareAddr // MAC地址
	state             bool             // true 为开启、 false为关闭
	mtu               int              // MTU
	logicFace         *LogicFace       // 对应的逻辑接口
	etherFaceMap      LogicFaceMap     // 对端mac地址和face对象映射表
	pcapHandle        *pcap.Handle     // pcap 抓包句柄
	receiveRoutineNum int
}

// Init
// @Description: 	初始化函数
// @receiver i
// @param name	网卡名
// @param macAddr	网卡mac地址
// @param mtu	网卡MTU
//
func (i *InterfaceListener) Init(name string, macAddr net.HardwareAddr, mtu int, receiveRoutineNum int) {
	i.name = name
	i.macAddr = macAddr
	i.mtu = mtu
	i.state = true
	i.receiveRoutineNum = receiveRoutineNum
}

// updateMtu 更新本网卡相关的所有 LogicFace 的MTU
//
// @Description:
// @receiver i
// @param mtu
//
func (i *InterfaceListener) updateMtu(mtu int) {
	i.logicFace.updateMTU(mtu)
	i.etherFaceMap.Range(func(key, value interface{}) bool {
		value.(*LogicFace).updateMTU(mtu)
		return true
	})
}

// Close
// @Description: 	关闭当前监听器，以及与该网卡相关的所有logicFace
// @receiver i
//
func (i *InterfaceListener) Close() {
	i.logicFace.Shutdown()
	i.etherFaceMap.Range(func(key, value interface{}) bool {
		lf := value.(*LogicFace)
		lf.Shutdown()
		return true
	})
}

// Start @Description: 启动当前监听器
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
	utils.GoroutineNoPanic(i.readPacketFromDev)
	return nil
}

// GetLogicFaceByMacAddr
// @Description: 	通过mac地址获得一个 LogicFace
// @receiver i
// @param macAddr
// @return *LogicFace
//
func (i *InterfaceListener) GetLogicFaceByMacAddr(macAddr string) *LogicFace {
	return i.etherFaceMap.LoadLogicFace(macAddr)
}

//
// @Description: 	收到包的处理函数
// @receiver i
// @param lpPacket	收到的包
// @param srcMacAddr	收到包的源MAC地址
//
func (i *InterfaceListener) onReceive(lpPacket *packet.LpPacket, srcMacAddr string) {
	logicFace := i.etherFaceMap.LoadLogicFace(srcMacAddr)
	if logicFace != nil {
		if logicFace.state == false { // 如果 logicface 已经关闭，则删除相应表项
			i.etherFaceMap.Delete(srcMacAddr)
		}
		logicFace.linkService.ReceivePacket(lpPacket)
		return
	}
	remoteMacAddr, err := net.ParseMAC(srcMacAddr)
	if err != nil {
		common2.LogWarn(err)
		return
	}
	if checkIdentity(lpPacket) == false {
		common2.LogInfo("user identify verify no pass")
		return
	}
	logicFacePtr, _ := createEtherLogicFace(i.name, i.macAddr, remoteMacAddr, i.mtu)
	if logicFacePtr == nil {
		common2.LogFatal("create ether logicface, error")
	}
	i.AddLogicFace(srcMacAddr, logicFacePtr)
	logicFacePtr.linkService.ReceivePacket(lpPacket)
}

//
// @Description: 处理收到的以太网帧协程
// @receiver i
//
func (i *InterfaceListener) processReceivedFrame(readPktChan <-chan gopacket.Packet) {
	for true {
		pkt, ok := <-readPktChan
		if !ok {
			common2.LogError("read from readPktChan error")
			break
		}
		lpPacketLen := binary.BigEndian.Uint16(pkt.Data()[14:16])
		lpPacket, err := parseByteArray2LpPacket(pkt.Data()[16 : 16+lpPacketLen])
		if err != nil {
			common2.LogError("parse byte to lpPacket error : ", err)
		} else {
			i.onReceive(lpPacket, pkt.LinkLayer().LinkFlow().Src().String())
		}
	}
}

//
// @Description: 	不断的从网卡中读包
// @receiver i
//
func (i *InterfaceListener) readPacketFromDev() {
	readPktChan := make(chan gopacket.Packet, 10000)
	common2.LogInfo("start interface: ", i.name, " receive routine number = ", i.receiveRoutineNum)
	for threadn := 0; threadn < i.receiveRoutineNum; threadn++ {
		utils.GoroutineNoPanic(func() {
			i.processReceivedFrame(readPktChan)
		})
	}

	pktSrc := gopacket.NewPacketSource(i.pcapHandle, i.pcapHandle.LinkType())
	for pkt := range pktSrc.Packets() {
		readPktChan <- pkt
	}
}

// DeleteLogicFace
// @Description: 通过mac地址删除一个etherFaceMap中的logicFace
// @receiver i
// @param remoteMacAddr
//
func (i *InterfaceListener) DeleteLogicFace(remoteMacAddr string) {
	i.etherFaceMap.Delete(remoteMacAddr)
}

// AddLogicFace
// @Description: 往etherFaceMap中添加一个 LogicFace
// @receiver i
// @param remoteMacAddr
// @param face
//
func (i *InterfaceListener) AddLogicFace(remoteMacAddr string, face *LogicFace) {
	i.etherFaceMap.StoreLogicFace(remoteMacAddr, face)
}
