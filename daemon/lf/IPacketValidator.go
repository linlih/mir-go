//
// @Author: Lihong Lin
// @Description:
// @Version: 1.0.0
// @Date: 2021/4/7 下午5:46
// @Copyright: MIN-Group；国家重大科技基础设施——未来网络北大实验室；深圳市信息论与未来网络重点实验室
//
package lf

type IPacketValidator interface {
	ReceiveMINPacket(data *IncomingPacketData)
}
