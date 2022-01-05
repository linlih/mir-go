// Package mgmt
// @Author: yzy
// @Description:
// @Version: 1.0.0
// @Date: 2021/3/10 11:37 上午
// @Copyright: MIN-Group；国家重大科技基础设施——未来网络北大实验室；深圳市信息论与未来网络重点实验室
//
package mgmt

import (
	"encoding/json"
	"fmt"
	"minlib/component"
	"minlib/mgmt"
	"minlib/packet"
)

// DataSender
// 发送数据包回调
//
// @Description:发送数据包回调
//
type DataSender func(data *packet.Data)

// DataSaver
// 保存数据包回调
//
// @Description:发送数据包回调
//
type DataSaver func(data *packet.Data)

// ResponseSender
// 发送错误信息回调
//
// @Description:发送错误信息回调
//
type ResponseSender func(response *mgmt.ControlResponse, interest *packet.Interest)

// StatusDatasetContext
// 分片数据集上下文结构体
//
// @Description:分片数据集上下文结构体
// @receiver s
//
type StatusDatasetContext struct {
	interest       *packet.Interest // 兴趣包指针
	FreshTime      uint64           // 生成的 data 的新鲜期，默认为 1 s
	items          []interface{}    // 数据
	SliceSize      int              // 每个分片的大小
	dataSender     DataSender       // 发送数据包回调
	responseSender ResponseSender   // 发送错误信息回调
	dataSaver      DataSaver        // 保存数据回调
}

// CreateSDC
// 创建数据集上下文函数
//
// @Description:创建数据集上下文函数
//
func CreateSDC(interest *packet.Interest, dataSender DataSender, nackSender ResponseSender, dataSaver DataSaver) *StatusDatasetContext {
	return &StatusDatasetContext{
		interest:       interest,
		FreshTime:      1000 * 1000,
		SliceSize:      7000,
		dataSender:     dataSender,
		responseSender: nackSender,
		dataSaver:      dataSaver,
	}
}

// Append 添加一条要发布的数据
//
// @Description:
// @receiver s
// @param item
//
func (s *StatusDatasetContext) Append(item interface{}) {
	s.items = append(s.items, item)
}

// AppendArray 添加一组要发布的数据
//
// @Description:
// @receiver s
// @param items
//
func (s *StatusDatasetContext) AppendArray(items []interface{}) {
	s.items = append(s.items, items...)
}

// Done 启动数据分片和缓存流程
//
// @Description:
// @receiver s
// @return uint64		返回分片数量
// @return error
//
func (s *StatusDatasetContext) Done(version uint64) error {
	data, err := json.Marshal(s.items)
	if err != nil {
		return err
	}
	// 分片数量
	var sliceNum = len(data)/s.SliceSize + 1
	// 如果正好整除的话 分片数减一
	if len(data)%s.SliceSize == 0 {
		sliceNum -= 1
	}
	currentIdentifier := s.interest.GetName()
	// 构造分片并缓存
	for i := 0; i < sliceNum; i++ {
		tempIdentifier, err := component.CreateIdentifierByComponents(currentIdentifier.GetComponents())
		if err != nil {
			s.Reject(MakeControlResponse(mgmt.ControlResponseCodeCommonError,
				fmt.Sprintf("Copy identifier failed => %v", err), ""))
		}
		tempIdentifier.AppendVersionNumber(version)
		tempIdentifier.AppendFragmentNumber(uint64(i))
		dataPacket := packet.NewDataByName(tempIdentifier)
		length := s.SliceSize
		if i == sliceNum-1 {
			length = len(data) - i*s.SliceSize
		}
		dataPacket.Payload.SetValue(data[i*s.SliceSize : i*s.SliceSize+length])
		s.dataSaver(dataPacket)
	}

	// 构造并返回元数据
	response, err := mgmt.CreateMetaDataControlResponse(version, uint64(sliceNum))
	if err != nil {
		s.Reject(MakeControlResponse(mgmt.ControlResponseCodeCommonError,
			fmt.Sprintf("Create meta response failed => %v", err), ""))
	}
	s.responseSender(response, s.interest)
	return nil
}

// Reject
// 如果在生成数据集的过程中发送了错误，可以通过本方法往用户侧发送一个表示错误的响应
//
// @Description:
// @receiver s
// @param response
//
func (s *StatusDatasetContext) Reject(response *mgmt.ControlResponse) {
	s.responseSender(response, s.interest)
}
