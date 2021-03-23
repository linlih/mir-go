//
// @Author: Jianming Que | weiguohua
// @Description:
// @Version: 1.0.0
// @Date: 2021/2/22 12:08 下午
// @Copyright: MIN-Group；国家重大科技基础设施——未来网络北大实验室；深圳市信息论与未来网络重点实验室
//
package lf

import "minlib/packet"

type LogicFaceType uint32

const (
	LogicFaceTypeTCP   LogicFaceType = 0
	LogicFaceTypeUDP   LogicFaceType = 1
	LogicFaceTypeEther LogicFaceType = 2
	LogicFaceTypeUnix  LogicFaceType = 3
)

//
// @Description: 逻辑接口类，用于发送网络分组，保存逻辑接口的状态信息等。
//
type LogicFace struct {
	LogicFaceId       uint64
	logicFaceType     LogicFaceType
	transport         ITransport
	linkService       *LinkService
	logicFaceCounters LogicFaceCounters
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
}

//
// @Description: 接收到包的处理函数，将包放入待处理缓冲区，更新统计数据
// @receiver lf
// @param minPacket
//
func (lf *LogicFace) ReceivePacket(minPacket *packet.MINPacket) {
	//TODO 把包入到待处理缓冲区
	//identifier, err := minPacket.GetIdentifier(0)

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
	lf.linkService.SendInterest(interest)
}

//
// @Description: 发送一个数据包
// @receiver lf
// @param data
//
func (lf *LogicFace) SendData(data *packet.Data) {
	lf.linkService.SendData(data)
}

//
// @Description: 发送一个Nack
// @receiver lf
// @param nack
//
func (lf *LogicFace) SendNack(nack *packet.Nack) {
	lf.linkService.SendNack(nack)
}

//
// @Description:  发送一个推送式包
// @receiver lf
// @param cPacket
//
func (lf *LogicFace) SendCPacket(cPacket *packet.CPacket) {
	lf.linkService.SendCPacket(cPacket)
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
	lf.transport.Close()
}
