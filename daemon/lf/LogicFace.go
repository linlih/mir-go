//
// @Author: Jianming Que
// @Description:
// @Version: 1.0.0
// @Date: 2021/2/22 12:08 下午
// @Copyright: MIN-Group；国家重大科技基础设施——未来网络北大实验室；深圳市信息论与未来网络重点实验室
//
package lf

import "minlib/packet"

type LogicFace struct {
	LogicFaceId uint64
}

func (lf *LogicFace) SendInterest(interest *packet.Interest) {

}

func (lf *LogicFace) SendData(data *packet.Data) {

}

func (lf *LogicFace) SendNack(nack *packet.Nack) {

}

func (lf *LogicFace) SendCPacket(cPacket *packet.CPacket) {

}
