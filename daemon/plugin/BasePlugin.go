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
// @Date: 2021/3/17 3:15 下午
// @Copyright: MIN-Group；国家重大科技基础设施——未来网络北大实验室；深圳市信息论与未来网络重点实验室
//
package plugin

import (
	"minlib/component"
	"minlib/packet"
	"mir-go/daemon/lf"
	"mir-go/daemon/table"
)

// BasePlugin
// 插件的基准实现，所有的锚点都返回默认的 0
//
// @Description:
//
type BasePlugin struct {
}

func (b BasePlugin) OnIncomingInterest(ingress *lf.LogicFace, interest *packet.Interest) int {
	return 0
}

func (b BasePlugin) OnInterestLoop(ingress *lf.LogicFace, interest *packet.Interest) int {
	return 0
}

func (b BasePlugin) OnContentStoreMiss(ingress *lf.LogicFace, pitEntry *table.PITEntry, interest *packet.Interest) int {
	return 0
}

func (b BasePlugin) OnContentStoreHit(ingress *lf.LogicFace, pitEntry *table.PITEntry, interest *packet.Interest, data *table.CSEntry) int {
	return 0
}

func (b BasePlugin) OnOutgoingInterest(egress *lf.LogicFace, pitEntry *table.PITEntry, interest *packet.Interest) int {
	return 0
}

func (b BasePlugin) OnInterestFinalize(pitEntry *table.PITEntry) int {
	return 0
}

func (b BasePlugin) OnIncomingData(ingress *lf.LogicFace, data *packet.Data) int {
	return 0
}

func (b BasePlugin) OnDataUnsolicited(ingress *lf.LogicFace, data *packet.Data) int {
	return 0
}

func (b BasePlugin) OnOutgoingData(egress *lf.LogicFace, data *packet.Data) int {
	return 0
}

func (b BasePlugin) OnIncomingNack(ingress *lf.LogicFace, nack *packet.Nack) int {
	return 0
}

func (b BasePlugin) OnOutgoingNack(egress *lf.LogicFace, pitEntry *table.PITEntry, header *component.NackHeader) int {
	return 0
}

func (b BasePlugin) OnIncomingGPPkt(ingress *lf.LogicFace, gPPkt *packet.GPPkt) int {
	return 0
}

func (b BasePlugin) OnOutgoingGPPkt(egress *lf.LogicFace, gPPkt *packet.GPPkt) int {
	return 0
}
