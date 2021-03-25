//
// @Author: weiguohua
// @Description:
// @Version: 1.0.0
// @Date: 2021/3/16 上午11:26
// @Copyright: MIN-Group；国家重大科技基础设施——未来网络北大实验室；深圳市信息论与未来网络重点实验室
//
package lf

//
// @Description: 统计信息对象，其关键成员有以下几个,用于统计一个logicFace的流量信息
//
type LogicFaceCounters struct {
	InCPacketN    uint64 // 从本接口流入的普通推式包的个数
	OutCPacketN   uint64 // 从本接口流出的普通推式包的个数
	DropCPacketN  uint64 // 从本接口流入后被丢弃的普通推式包的个数
	InInterestN   uint64 // 从本接口流入的兴趣包的个数
	OutInterestN  uint64 // 从本接口流出的兴趣包的个数
	DropInterestN uint64 // 从本接口流入后被丢弃的兴趣包的个数
	InDataN       uint64 // 从本接口流入的数据包的个数
	OutDataN      uint64 // 从本接口流出的数据包的个数
	DropDataN     uint64 // 从本接口流入后被丢弃的数据包的个数
	InNackN       uint64 // 从本接口流入的Nack包的个数
	OutNackN      uint64 // 从本接口流出的Nack包的个数
	DropNackN     uint64 // 从本接口流入后被丢弃的Nack包的个数
	InBytesN      uint64 // 从本接口流入的数据字节数
	OutBytesN     uint64 // 从本接口流出的数据字节数
}