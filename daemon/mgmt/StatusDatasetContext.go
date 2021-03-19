//
// @Author: Jianming Que
// @Description:
// @Version: 1.0.0
// @Date: 2021/3/10 11:37 上午
// @Copyright: MIN-Group；国家重大科技基础设施——未来网络北大实验室；深圳市信息论与未来网络重点实验室
//
package mgmt

import (
	"fmt"
	"minlib/component"
	"minlib/encoding"
	"minlib/mgmt"
	"minlib/packet"
	"mir-go/daemon/utils"
	"time"
)

const (
	INITIAL = iota 	///< none of .append, .end, .reject has been invoked
	RESPONDED 		///< .append has been invoked
	FINALIZED 		///< .end or .reject has been invoked
)

//
// 数据集上下文，由 Dispatcher 创建，传递给具体的管理模块，管理模块调用本上下文对象将数据通过 Dispatcher 发出
//
// @Description:
//

type DataSender func(component.Identifier,*encoding.Block,time.Duration,bool)

type NackSender	func(*mgmt.ControlResponse)

type StatusDatasetContext struct {
	Prefix    			component.Identifier // 要发布的数据的前缀
	FreshTime 			time.Duration         // 生成的 Data 的新鲜期，默认为 1 s
	state				int
	segmentNo   		uint64
	encodingBuffer		*encoding.Encoder
	interest			*packet.Interest
	dataSender			DataSender
	nackSender			NackSender
}

 func CreateSDC(interest *packet.Interest,dataSender DataSender,nackSender NackSender)*StatusDatasetContext{
 	return &StatusDatasetContext{
 		Prefix: *interest.GetName(),
 		state: INITIAL,
		FreshTime: 100*time.Millisecond,
 		interest: interest,
 		dataSender: dataSender,
 		nackSender: nackSender,
	}
 }


//
// 添加一个要发送的 Block 作为响应
//
// @Description:
// @receiver s
// @param block
//
func (s *StatusDatasetContext) Append(block *encoding.Block) {
	if  s.state == FINALIZED{
		fmt.Errorf("state is in FINALIZED")
		return
	}
	// 依次发送块 最后一块做特殊处理表示结束
	s.state = RESPONDED
	size,_:=block.Size()
	byteArrLeft := size
	for byteArrLeft>0{
		// 延迟初始化
		if s.encodingBuffer == nil{
			s.encodingBuffer = &encoding.Encoder{}
			s.encodingBuffer.EncoderReset(encoding.MaxPacketSize, 0)
		}
		// 每次发送的数据大小为 MaxPacketSize
		// 如果block里面的数据量 小于这个数值
		// 直接发送 如果大于 发送最大值
		nBytesAppend:=utils.Min(byteArrLeft,encoding.MaxPacketSize)
		s.encodingBuffer.AppendByteArray(block.GetRaw()[size-byteArrLeft:],nBytesAppend)
		byteArrLeft -= nBytesAppend
		if byteArrLeft>0{
			dataPrefix:=s.Prefix
			dataPrefix.Append(component.CreateIdentifierComponentByNonNegativeInteger(s.segmentNo))
			s.segmentNo+=1
			// 该函数需要修改 要等待函数实现 应该根据encoder创建一个新的block 发送过去
			s.dataSender(dataPrefix,block,s.FreshTime,false)
			// 重置
			s.encodingBuffer.EncoderReset(encoding.MaxPacketSize, 0)
		}
	}
}

//
// 所有数据 Append 完毕之后，调用本方法，会生成一个可以标识数据集结束的 Data ，用户侧收到这种特殊的包即可判定本次数据拉取结束
//
// @Description:
// @receiver s
// @param block
//
func (s *StatusDatasetContext) Finish() {
	if s.state == FINALIZED{
		fmt.Errorf("state is in FINALIZED")
		return
	}
	s.state = FINALIZED
	dataPrefix:=s.Prefix
	dataPrefix.Append(component.CreateIdentifierComponentByNonNegativeInteger(s.segmentNo))
	// 发送最后一个分片过去 最后一个参数指示是否为最后一个分片
	s.dataSender(dataPrefix,block,s.FreshTime,true)
}

//
// 如果在生成数据集的过程中发送了错误，可以通过本方法往用户侧发送一个表示错误的响应
//
// @Description:
// @receiver s
// @param response
//
func (s *StatusDatasetContext) Reject(response *mgmt.ControlResponse) {
	if s.state != INITIAL{
		fmt.Errorf("state is in RESPONDED or FINALIZED")
		return
	}
	s.state = FINALIZED
	s.nackSender(response)
}
