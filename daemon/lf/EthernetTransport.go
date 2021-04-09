//
// @Author: weiguohua
// @Description:
// @Version: 1.0.0
// @Date: 2021/3/16 上午11:36
// @Copyright: MIN-Group；国家重大科技基础设施——未来网络北大实验室；深圳市信息论与未来网络重点实验室
//
package lf

import (
	"github.com/google/gopacket/pcap"
	"minlib/packet"
	"mir-go/daemon/common"
	"net"
	"time"
)

//
// @Description:  用来发送和接收以太网帧
//   |---- 8字节目的Mac地址----|---- 8字节源Mac地址----|-2字节协议号-|-LpPacket-|
//
type EthernetTransport struct {
	Transport
	deviceName    string
	localMacAddr  net.HardwareAddr // MAC地址
	remoteMacAddr net.HardwareAddr // MAC地址
	snapshotLen   int32            // 抓包长度
	promiscuous   bool             // 混杂模式
	timeout       time.Duration    // 超时时间 <= 0表示不超时
	handle        *pcap.Handle     // 文件描述符
	status        bool             // 状态
	sendPacket    [10000]byte      // 发包缓冲区

}

//
// @Description: 初始化本对象
// @receiver e
// @param ifName	网卡名称
// @param localMacAddr	本地Mac地址
// @param remoteMacAddr	对端Mac地址
// @param etherTransportMap	对端mac地址和face对象映射表，通过以太网接收到以太网帧时，通过以太网帧的源MAC地址，
//			在 etherTransportMap 查找，确定用哪个logicFace来处理收到的包
//
func (e *EthernetTransport) Init(ifName string, localMacAddr, remoteMacAddr net.HardwareAddr) {
	e.deviceName = ifName
	e.snapshotLen = 10240 // 抓包的大小
	e.promiscuous = false // 混杂模式
	e.timeout = -1        // 超时时间

	e.localMacAddr = localMacAddr
	e.remoteMacAddr = remoteMacAddr

	e.localAddr = localMacAddr.String()
	e.remoteAddr = remoteMacAddr.String()
	e.localUri = "ether://" + e.localAddr
	e.remoteUri = "ether://" + e.remoteAddr

	//e.etherTransportMap = etherTransportMap
	// 设置以太网包头部
	copy(e.sendPacket[0:6], remoteMacAddr)
	copy(e.sendPacket[6:12], localMacAddr)
	e.sendPacket[12] = 0x88
	e.sendPacket[13] = 0x88

	e.status = true
	var err error
	e.handle, err = pcap.OpenLive(e.deviceName, e.snapshotLen, e.promiscuous, e.timeout)
	if err != nil {
		e.status = false
		common.LogError("open default net device error", err)
		return
	}
	//mPcapFilter := "ether proto 0x8600"
	common.LogInfo(e.localAddr)
	err = e.handle.SetBPFFilter("ether proto 0x8888 and not ether src host " + e.localAddr)
	if err != nil {
		e.status = false
		common.LogError("open default net device error", err)
		return
	}
}

//
// @Description: 	设置linkService
// @receiver e
// @param service	linkService对象指针
//
func (e *EthernetTransport) SetLinkService(service *LinkService) {
	e.linkService = service
}

//
// @Description: 	关闭
// @receiver e
//
func (e *EthernetTransport) Close() {
	e.handle.Close()
}

// //
// @Description: 发送以太网包
// @receiver e
// @param lpPacket	以太网包对象
//
func (e *EthernetTransport) Send(lpPacket *packet.LpPacket) {
	encodeBufLen, encodeBuf := encodeLpPacket2ByteArray(lpPacket)
	if encodeBufLen <= 0 {
		return
	}
	copy(e.sendPacket[14:], encodeBuf[0:encodeBufLen])
	err := e.handle.WritePacketData(e.sendPacket[0 : 14+encodeBufLen])
	if err != nil {
		common.LogWarn(err, ", packet len = ", 14+encodeBufLen)
	}
}

//
// @Description: 收到包后调用linkService处理
// @receiver e
// @param lpPacket	收到的包
//
func (e *EthernetTransport) onReceive(lpPacket *packet.LpPacket, srcMacAddr string) {
	// TODO 暂时不用这个函数
}

//
// @Description: 	阻塞读，由上层线程或协程调用
// @receiver e
//
func (e *EthernetTransport) Receive() {
	// TODO 暂时不用这个函数
}
