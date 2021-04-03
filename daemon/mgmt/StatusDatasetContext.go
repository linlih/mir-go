//
// @Author: yzy
// @Description:
// @Version: 1.0.0
// @Date: 2021/3/10 11:37 上午
// @Copyright: MIN-Group；国家重大科技基础设施——未来网络北大实验室；深圳市信息论与未来网络重点实验室
//
package mgmt

import (
	"minlib/component"
	"minlib/encoding"
	"minlib/mgmt"
	"minlib/packet"
	"mir-go/daemon/common"
	"mir-go/daemon/utils"
	"strconv"
	"time"
)

const (
	INITIAL = iota
	RESPONDED
	FINALIZED
)

//
// 发送数据包回调
//
// @Description:发送数据包回调
//
type DataSender func(data *packet.Data)

//
// 发送错误信息回调
//
// @Description:发送错误信息回调
//
type NackSender func(response *mgmt.ControlResponse, interest *packet.Interest)

//
// 分片数据集上下文结构体
//
// @Description:分片数据集上下文结构体
// @receiver s
//
type StatusDatasetContext struct {
	interest   *packet.Interest     // 兴趣包指针
	Prefix     component.Identifier // 要发布的数据的前缀
	FreshTime  time.Duration        // 生成的 Data 的新鲜期，默认为 1 s
	state      int                  // 兴趣包状态
	segmentNo  int                  // 分片号
	data       []byte               // 分片数据
	dataSender DataSender           // 发送数据包回调
	nackSender NackSender           // 发送错误信息回调
}

//
// 创建数据集上下文函数
//
// @Description:创建数据集上下文函数
//
func CreateSDC(interest *packet.Interest, dataSender DataSender, nackSender NackSender) *StatusDatasetContext {
	return &StatusDatasetContext{
		Prefix:     *interest.GetName(),
		state:      INITIAL,
		FreshTime:  100 * time.Millisecond,
		dataSender: dataSender,
		nackSender: nackSender,
	}
}

//
// 对数据集分片并缓存和发送
//
// @Description:
// @receiver s
//
func (s *StatusDatasetContext) Append() {
	if s.state == FINALIZED {
		common.LogWarn("state is in FINALIZED")
		return
	}
	s.state = RESPONDED
	size := encoding.SizeT(len(s.data))
	byteArrLeft := size
	for byteArrLeft > 0 {
		nBytesAppend := utils.Min(byteArrLeft, encoding.MaxPacketSize)
		data := &packet.Data{}
		// 从1开始是分片
		s.segmentNo += 1
		//解引用防止篡改源数据
		prefix := s.Prefix
		prefix.Append(component.CreateIdentifierComponentByNonNegativeInteger(uint64(s.segmentNo)))
		data.SetName(&prefix)
		data.Payload.SetValue(s.data[size-byteArrLeft : size-byteArrLeft+nBytesAppend])

		byteArrLeft -= nBytesAppend
		if byteArrLeft <= 0 {
			s.state = FINALIZED
		}
		dispatcher.Cache.Add(s.Prefix.ToUri()+strconv.Itoa(s.segmentNo), data)
		s.dataSender(data)
	}
}

//
// 如果在生成数据集的过程中发送了错误，可以通过本方法往用户侧发送一个表示错误的响应
//
// @Description:
// @receiver s
// @param response
//
func (s *StatusDatasetContext) Reject(response *mgmt.ControlResponse) {
	if s.state != INITIAL {
		common.LogWarn("state is in RESPONDED or FINALIZED")
		return
	}
	s.state = FINALIZED
	s.nackSender(response, s.interest)
}
