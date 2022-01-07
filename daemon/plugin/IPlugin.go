// Copyright [2022] [MIN-Group -- Peking University Shenzhen Graduate School Multi-Identifier Network Development Group]
//
// Licensed under the Apache License, Version 2.0 (the "License"): you may
// not use this file except in compliance with the License. You may obtain
// a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS, WITHOUT
// WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the
// License for the specific language governing permissions and limitations
// under the License.

// Package plugin
// @Author: Jianming Que
// @Description:
// @Version: 1.0.0
// @Date: 2021/3/17 2:35 下午
// @Copyright: MIN-Group；国家重大科技基础设施——未来网络北大实验室；深圳市信息论与未来网络重点实验室
//
package plugin

import (
	"minlib/component"
	"minlib/packet"
	"mir-go/daemon/lf"
	"mir-go/daemon/table"
)

// IPlugin
// 插件接口，所有的插件都应该实现以下接口
//
// @Description:
//
type IPlugin interface {
	// OnIncomingInterest
	// Incoming Interest 管道锚点
	//
	// @Description:
	// @param ingress
	// @param interest
	// @return int	0 => 继续执行
	//			    -1 => 拦截执行
	//
	OnIncomingInterest(ingress *lf.LogicFace, interest *packet.Interest) int

	// OnInterestLoop
	// Interest Loop 管道锚点
	//
	// @Description:
	// @param ingress
	// @param interest
	// @return int	0 => 继续执行
	//			    -1 => 拦截执行
	//
	OnInterestLoop(ingress *lf.LogicFace, interest *packet.Interest) int

	// OnContentStoreMiss
	// ContentStore miss 管道锚点
	//
	// @Description:
	// @param ingress
	// @param pitEntry
	// @param interest
	// @return int	0 => 继续执行
	//			    -1 => 拦截执行
	//
	OnContentStoreMiss(ingress *lf.LogicFace, pitEntry *table.PITEntry, interest *packet.Interest) int

	// OnContentStoreHit
	// ContentStore hit 管道锚点
	//
	// @Description:
	// @param ingress
	// @param pitEntry
	// @param interest
	// @param data
	// @return int
	//
	OnContentStoreHit(ingress *lf.LogicFace, pitEntry *table.PITEntry, interest *packet.Interest, data *table.CSEntry) int

	// OnOutgoingInterest
	// Outgoing Interest 管道锚点
	//
	// @Description:
	// @param egress
	// @param pitEntry
	// @param interest
	// @return int
	//
	OnOutgoingInterest(egress *lf.LogicFace, pitEntry *table.PITEntry, interest *packet.Interest) int

	// OnInterestFinalize
	// Interest Finalize 管道锚点
	//
	// @Description:
	// @param pitEntry
	// @return int
	//
	OnInterestFinalize(pitEntry *table.PITEntry) int

	// OnIncomingData
	// Incoming data 管道锚点
	//
	// @Description:
	// @param ingress
	// @param data
	// @return int
	//
	OnIncomingData(ingress *lf.LogicFace, data *packet.Data) int

	// OnDataUnsolicited
	// data unsolicited 管道锚点
	//
	// @Description:
	// @param ingress
	// @param data
	// @return int
	//
	OnDataUnsolicited(ingress *lf.LogicFace, data *packet.Data) int

	// OnOutgoingData
	// Outgoing data 管道锚点
	//
	// @Description:
	// @param egress
	// @param data
	// @return int
	//
	OnOutgoingData(egress *lf.LogicFace, data *packet.Data) int

	// OnIncomingNack
	// Incoming Nack 管道锚点
	//
	// @Description:
	// @param ingress
	// @param nack
	// @return int
	//
	OnIncomingNack(ingress *lf.LogicFace, nack *packet.Nack) int

	// OnOutgoingNack
	// Outgoing Nack 管道锚点
	//
	// @Description:
	// @param egress
	// @param pitEntry
	// @param header
	// @return int
	//
	OnOutgoingNack(egress *lf.LogicFace, pitEntry *table.PITEntry, header *component.NackHeader) int

	// OnIncomingGPPkt
	// Incoming GPPkt 管道锚点
	//
	// @Description:
	// @param ingress
	// @param gPPkt
	// @return int
	//
	OnIncomingGPPkt(ingress *lf.LogicFace, gPPkt *packet.GPPkt) int

	// OnOutgoingGPPkt
	// Outgoing GPPkt 管道锚点
	//
	// @Description:
	// @param egress
	// @param gPPkt
	// @return int
	//
	OnOutgoingGPPkt(egress *lf.LogicFace, gPPkt *packet.GPPkt) int
}
