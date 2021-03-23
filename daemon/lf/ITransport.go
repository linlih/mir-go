//
// @Author: weiguohua
// @Description:
// @Version: 1.0.0
// @Date: 2021/3/17 下午6:03
// @Copyright: MIN-Group；国家重大科技基础设施——未来网络北大实验室；深圳市信息论与未来网络重点实验室
//
package lf

import "minlib/packet"

//
// @Description:  Tranport 接口， 便于LogicFace声明成员
//
type ITransport interface {
	//
	// @Description:  关闭
	//
	Close()
	//
	// @Description: 发送一个lpPacket
	// @param lpPacket
	//
	Send(lpPacket *packet.LpPacket)
	//
	// @Description: 从网络中接收到一段数据
	//
	Receive()
	//
	// @Description: 获得Transport的对端地址
	// @return string	对端地址
	//
	GetRemoteUri() string
	//
	// @Description: 获得Transport的本机地址
	// @return string	本机地址
	//
	GetLocalUri() string
}
