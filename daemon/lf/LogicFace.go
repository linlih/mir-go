//
// @Author: Jianming Que | weiguohua
// @Description:
// @Version: 1.0.0
// @Date: 2021/2/22 12:08 下午
// @Copyright: MIN-Group；国家重大科技基础设施——未来网络北大实验室；深圳市信息论与未来网络重点实验室
//
package lf

import (
	"log"
	"minlib/encoding"
	"minlib/packet"
)

type LogicFaceType uint32

const (
	LogicFaceTypeTCP   LogicFaceType = 0
	LogicFaceTypeUDP   LogicFaceType = 1
	LogicFaceTypeEther LogicFaceType = 2
	LogicFaceTypeUnix  LogicFaceType = 3
)

const MaxIdolTimeMs = 600

//
// @Description: 逻辑接口类，用于发送网络分组，保存逻辑接口的状态信息等。
//
type LogicFace struct {
	LogicFaceId       uint64
	logicFaceType     LogicFaceType
	transport         ITransport
	linkService       *LinkService
	logicFaceCounters LogicFaceCounters
	expireTime        int64 // 超时时间 ms
	state             bool  //  true 为 up , false 为down
}

//
// @Description: 	初始化logicFace
// @receiver lf
// @param transport	transport对象指针
// @param linkService
//
func (lf *LogicFace) Init(transport ITransport, linkService *LinkService, faceType LogicFaceType) {
	lf.transport = transport
	lf.linkService = linkService
	lf.logicFaceType = faceType
	lf.state = true
	lf.expireTime = getTimestampMS() + MaxIdolTimeMs
}

//
// @Description: 接收到包的处理函数，将包放入待处理缓冲区，更新统计数据
// @receiver lf
// @param minPacket
//
func (lf *LogicFace) ReceivePacket(minPacket *packet.MINPacket) {
	//TODO 把包入到待处理缓冲区
	identifier, err := minPacket.GetIdentifier(0)
	if err != nil {
		log.Println("face ", lf.LogicFaceId, " receive packet has no identifier")
		return
	}
	if identifier.GetIdentifierType() == encoding.TlvIdentifierCommon {
		lf.logicFaceCounters.InCPacketN++
	} else if identifier.GetIdentifierType() == encoding.TlvIdentifierContentInterest {
		lf.logicFaceCounters.InInterestN++
	} else if identifier.GetIdentifierType() == encoding.TlvIdentifierContentData {
		lf.logicFaceCounters.InDataN++
	}

	lf.expireTime = getTimestampMS() + MaxIdolTimeMs

}

//
// @Description: 	启动接收数据协程
// @receiver lf
//
func (lf *LogicFace) Start() {
	// 启动收包协程
	go lf.transport.Receive()
}

//
// @Description:  发送一个MIN包
// @receiver lf
// @param packet
//
func (lf *LogicFace) SendMINPacket(packet *packet.MINPacket) {

}

//
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

//
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

//
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

//
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

//
// @Description: 获得本地地址
// @receiver lf
// @return string
//
func (lf *LogicFace) GetLocalUri() string {
	return lf.transport.GetLocalUri()
}

//
// @Description: 获得对端地址
// @receiver lf
// @return string
//
func (lf *LogicFace) GetRemoteUri() string {
	return lf.transport.GetRemoteUri()
}

//
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
