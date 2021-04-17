// Package lf
// @Author: Jianming Que | weiguohua
// @Description:
// @Version: 1.0.0
// @Date: 2021/2/22 12:08 下午
// @Copyright: MIN-Group；国家重大科技基础设施——未来网络北大实验室；深圳市信息论与未来网络重点实验室
//
package lf

import (
	common2 "minlib/common"
	"minlib/encoding"
	"minlib/packet"
	"sync"
)

type LogicFaceType uint32

var lock sync.Mutex

//
// @Description:  LogicFace的类型
//
const (
	LogicFaceTypeTCP   LogicFaceType = 0
	LogicFaceTypeUDP   LogicFaceType = 1
	LogicFaceTypeEther LogicFaceType = 2
	LogicFaceTypeUnix  LogicFaceType = 3
	LogicFaceTypeInner LogicFaceType = 4
)

// MaxIdolTimeMs
// @Description:  超过 600s 没有接收数据或发送数据的logicFace会被logicFaceSystem的face cleaner销毁
//
const MaxIdolTimeMs = 600000

// LogicFace
// @Description: 逻辑接口类，用于发送网络分组，保存逻辑接口的状态信息等。
//		LogicFace-LinkService-Transport是一个 一一对应的关系，他们相互绑定
//		在一个收包流程中网络数据最开始是通过transport流入的，由transport调用LinkService的 receive函数处理接收到的网络包，
//		再由linkService调用logicFace的receive函数。
//		在一个发送包的流程中，由logicFace调用linkService的发包函数，再由linkService调用transport的发包函数
//
type LogicFace struct {
	LogicFaceId       uint64 // logicFaceID
	logicFaceType     LogicFaceType
	transport         ITransport        // 与logicFace绑定的transport
	linkService       *LinkService      // 与logicFace绑定的linkService
	logicFaceCounters LogicFaceCounters // logicFace 流量统计对象
	expireTime        int64             // 超时时间 ms
	state             bool              //  true 为 up , false 为down
}

// Init
// @Description: 	初始化logicFace
// @receiver lf
// @param transport	transport对象指针
// @param linkService
// @param faceType   face类型
//
func (lf *LogicFace) Init(transport ITransport, linkService *LinkService, faceType LogicFaceType) {
	lf.transport = transport
	lf.linkService = linkService
	lf.logicFaceType = faceType
	lf.state = true
	lf.expireTime = getTimestampMS() + MaxIdolTimeMs
}

// ReceivePacket
// @Description: 接收到包的处理函数，将包放入待处理缓冲区，更新统计数据
// @receiver lf
// @param minPacket
//
func (lf *LogicFace) ReceivePacket(minPacket *packet.MINPacket) {
	//common2.LogInfo("receive packet from logicFace : ", lf.LogicFaceId, " ", lf.GetRemoteUri())
	//把包入到待处理缓冲区
	gLogicFaceSystem.packetValidator.ReceiveMINPacket(&IncomingPacketData{
		LogicFace: lf,
		MinPacket: minPacket,
	})
	identifier, err := minPacket.GetIdentifier(0)
	if err != nil {
		common2.LogWarn(err, "face ", lf.LogicFaceId, " receive packet has no identifier")
		return
	}
	if identifier.GetIdentifierType() == encoding.TlvIdentifierCommon {
		lf.logicFaceCounters.InCPacketN++
	} else if identifier.GetIdentifierType() == encoding.TlvIdentifierContentInterest {
		lock.Lock()
		lf.logicFaceCounters.InInterestN++
		lock.Unlock()
		//common2.LogInfo(lf.logicFaceCounters.InInterestN)
	} else if identifier.GetIdentifierType() == encoding.TlvIdentifierContentData {
		lf.logicFaceCounters.InDataN++
	}

	lf.expireTime = getTimestampMS() + MaxIdolTimeMs

}

// Start
// @Description: 	启动接收数据协程
// @receiver lf
//
func (lf *LogicFace) Start() {
	// 启动收包协程
	go lf.transport.Receive()
}

// SendMINPacket
// @Description:  发送一个MIN包
// @receiver lf
// @param packet
//
func (lf *LogicFace) SendMINPacket(packet *packet.MINPacket) {

}

// SendInterest
// @Description: 发送一个兴趣包
// @receiver lf
// @param interest
//
func (lf *LogicFace) SendInterest(interest *packet.Interest) {
	if !lf.state {
		return
	}
	lf.linkService.SendInterest(interest)
	lf.expireTime = getTimestampMS() + MaxIdolTimeMs
}

// SendData
// @Description: 发送一个数据包
// @receiver lf
// @param data
//
func (lf *LogicFace) SendData(data *packet.Data) {
	if !lf.state {
		return
	}
	lf.linkService.SendData(data)
	lf.expireTime = getTimestampMS() + MaxIdolTimeMs
}

// SendNack
// @Description: 发送一个Nack
// @receiver lf
// @param nack
//
func (lf *LogicFace) SendNack(nack *packet.Nack) {
	if !lf.state {
		return
	}
	lf.linkService.SendNack(nack)
	lf.expireTime = getTimestampMS() + MaxIdolTimeMs
}

// SendCPacket
// @Description:  发送一个推送式包
// @receiver lf
// @param cPacket
//
func (lf *LogicFace) SendCPacket(cPacket *packet.CPacket) {
	if !lf.state {
		return
	}
	lf.linkService.SendCPacket(cPacket)
	lf.expireTime = getTimestampMS() + MaxIdolTimeMs
}

// GetLocalUri
// @Description: 获得本地地址
// @receiver lf
// @return string
//
func (lf *LogicFace) GetLocalUri() string {
	return lf.transport.GetLocalUri()
}

// GetRemoteUri
// @Description: 获得对端地址
// @receiver lf
// @return string
//
func (lf *LogicFace) GetRemoteUri() string {
	return lf.transport.GetRemoteUri()
}

// Shutdown
// @Description: 关闭face
// @receiver lf
//
func (lf *LogicFace) Shutdown() {

	if lf.state == false {
		return
	}

	if lf.logicFaceType != LogicFaceTypeUDP {
		lf.transport.Close()
	}
	lf.state = false
}

func (lf *LogicFace) GetCounter() uint64 {
	return lf.logicFaceCounters.InInterestN
}
