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
var logicFaceMaxIdolTimeMs int64 = 600000

//var logicFaceMaxIdolTimeMs int64 = 5000

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
	state             bool              //  true 为 up , false 为 down
	Mtu               uint64            // 最大传输单元 MTU
	Persistence       uint64            // 持久性, 0 表示没有持久性，会被LogicFaceSystem在一定时间后清理掉
	//	非 0 时表示有持久性，就算一直没有收发数据，也不会被清理
	onShutdownCallback func(logicFaceId uint64) // 传输logic face 关闭时的回调

	sendQue chan encoding.IEncodingAble
	recvQue chan *packet.MINPacket
}

// GetState 获取接口状态
//
// @Description:
// @receiver lf
// @return bool
//
func (lf *LogicFace) GetState() bool {
	return lf.state
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
	lf.expireTime = getTimestampMS() + logicFaceMaxIdolTimeMs
	lf.Mtu = uint64(linkService.mtu)
	lf.Persistence = 0

	lf.recvQue = make(chan *packet.MINPacket, gLogicFaceSystem.config.LFRecvQueSize)
	lf.sendQue = make(chan encoding.IEncodingAble, gLogicFaceSystem.config.LFSendQueSize)
}

//
// @Description: 用于处理给已关闭的chan发数据异常
// @param oldErr
// @return {
//
var send2ChanException = func() {
	if r := recover(); r != nil && r.(error).Error() == "send on closed channel" {
		common2.LogError("write to chan error: send on closed channel")
	}
}

// ReceivePacket
// @Description: 接收到包的处理函数，将包放入接收队列，如果队列满了，则丢包
// @receiver lf
// @param minPacket
//
func (lf *LogicFace) ReceivePacket(minPacket *packet.MINPacket) {
	defer send2ChanException()
	if !lf.state {
		return
	}
	if len(lf.recvQue) < cap(lf.recvQue) {
		lf.recvQue <- minPacket
	} else {
		common2.LogDebug("receive que full, ", lf.GetLocalUri(), lf.GetRemoteUri())
	}
}

//
// @Description:	由接收协程调用，把接收队列中的包往forwarder的缓冲区中送
// @receiver lf
// @param minPacket
//
func (lf *LogicFace) onReceivePacket(minPacket *packet.MINPacket) {
	common2.LogDebug("receive packet from logicFace : ", lf.LogicFaceId, " ", lf.GetRemoteUri())
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
		lf.logicFaceCounters.InGPPktN++
	} else if identifier.GetIdentifierType() == encoding.TlvIdentifierContentInterest {
		lock.Lock()
		lf.logicFaceCounters.InInterestN++
		lock.Unlock()
		//common2.LogInfo(lf.logicFaceCounters.InInterestN)
	} else if identifier.GetIdentifierType() == encoding.TlvIdentifierContentData {
		lf.logicFaceCounters.InDataN++
	}

	lf.expireTime = getTimestampMS() + logicFaceMaxIdolTimeMs
}

// Start
// @Description: 	启动接收数据协程
// @receiver lf
//
func (lf *LogicFace) Start() {
	// 启动收包协程
	go lf.transport.Receive()
	// 启动收包协程，负责把logic face 收到的包往forwarder的队列送
	go func() {
		for lf.state {
			minPacket, ok := <-lf.recvQue
			if !ok {
				common2.LogError("read packet from recv que error")
				lf.Shutdown()
				break
			}
			lf.onReceivePacket(minPacket)
		}
	}()
	// 启动发包协程，负责把forwarder 发往该 logic face 的包转发出去
	go func() {
		for lf.state {
			minPacket, ok := <-lf.sendQue
			if !ok {
				common2.LogError("read packet from send que error")
				lf.Shutdown()
				break
			}
			lf.linkService.SendEncodingAble(minPacket)
		}
	}()
}

func (lf *LogicFace) addPkt2SendQue(pkt encoding.IEncodingAble) {
	defer send2ChanException()
	if !lf.state {
		return
	}
	if len(lf.sendQue) < cap(lf.sendQue) {
		lf.sendQue <- pkt
	}
}

// SendMINPacket
// @Description:  发送一个MIN包
// @receiver lf
// @param packet
//
func (lf *LogicFace) SendMINPacket(packet *packet.MINPacket) {
	lf.addPkt2SendQue(packet)
}

// SendInterest
// @Description: 发送一个兴趣包
// @receiver lf
// @param interest
//
func (lf *LogicFace) SendInterest(interest *packet.Interest) {
	lf.addPkt2SendQue(interest)
}

// SendData
// @Description: 发送一个数据包
// @receiver lf
// @param data
//
func (lf *LogicFace) SendData(data *packet.Data) {
	lf.addPkt2SendQue(data)
}

// SendNack
// @Description: 发送一个Nack
// @receiver lf
// @param nack
//
func (lf *LogicFace) SendNack(nack *packet.Nack) {
	lf.addPkt2SendQue(nack)
}

// SendGPPkt
// @Description:  发送一个推送式包
// @receiver lf
// @param gPPkt
//
func (lf *LogicFace) SendGPPkt(gPPkt *packet.GPPkt) {
	lf.addPkt2SendQue(gPPkt)
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
	lf.state = false
	close(lf.sendQue)
	close(lf.recvQue)
	lf.transport.Close()

	common2.LogInfo("logic face : ", lf.LogicFaceId, " is shutdown")
	lf.onLogicFaceShutDown()
}

func (lf *LogicFace) GetCounter() uint64 {
	return lf.logicFaceCounters.InInterestN
}

//
// @Description: 	设置LogicFace的Persistence 属性，当persistence 不为0是， 该logicFace不会因为长时间不用被删除
// @receiver lf
// @param persistence
//
func (lf *LogicFace) SetPersistence(persistence uint64) {
	lf.Persistence = persistence
}

func (lf *LogicFace) onLogicFaceShutDown() {
	lf.state = false
	if lf.onShutdownCallback != nil {
		lf.onShutdownCallback(lf.LogicFaceId)
	}
}

func (lf *LogicFace) SetOnShutdownCallback(callback func(logicFaceId uint64)) {
	lf.onShutdownCallback = callback
}
