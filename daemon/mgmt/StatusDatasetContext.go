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
	"strconv"
	"time"
)

const (
	INITIAL = iota
	RESPONDED
	FINALIZED
)

//
// 数据集上下文，由 Dispatcher 创建，传递给具体的管理模块，管理模块调用本上下文对象将数据通过 Dispatcher 发出
//
// @Description:
//

type DataSender func(data *packet.Data)

type NackSender func(response *mgmt.ControlResponse, interest *packet.Interest)

type StatusDatasetContext struct {
	interest   *packet.Interest
	Prefix     component.Identifier // 要发布的数据的前缀
	FreshTime  time.Duration        // 生成的 Data 的新鲜期，默认为 1 s
	state      int
	segmentNo  int
	data       []byte
	dataSender DataSender
	nackSender NackSender
}

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
// 添加一个要发送的 Block 作为响应
//
// @Description:
// @receiver s
// @param block
//
func (s *StatusDatasetContext) Append() {
	if s.state == FINALIZED {
		fmt.Println("state is in FINALIZED")
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
		fmt.Println("state is in RESPONDED or FINALIZED")
		return
	}
	s.state = FINALIZED
	s.nackSender(response, s.interest)
}
