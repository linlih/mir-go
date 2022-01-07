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

// Package lf
// @Author: Jianming Que
// @Description:
// @Version: 1.0.0
// @Date: 2021/3/26 11:24 上午
// @Copyright: MIN-Group；国家重大科技基础设施——未来网络北大实验室；深圳市信息论与未来网络重点实验室
//
package lf

import (
	"github.com/sirupsen/logrus"
	"minlib/packet"
)

// IncomingPacketData
// 包装了一个 LogicFace 和 MINPacket 的指针，主要用于 LogicFace 和 Forwarder 进行通信
//
// @Description:
//	1. 主要目的是告诉 Forwarder 从哪个 LogicFace 收到了一个网络包
//
type IncomingPacketData struct {
	LogicFace *LogicFace
	MinPacket *packet.MINPacket
}

func (ipd *IncomingPacketData) ToFields() logrus.Fields {
	firstIdentifier, err := ipd.MinPacket.GetIdentifier(0)
	if err != nil {
		return logrus.Fields{
			"LogicFace": ipd.LogicFace.LogicFaceId,
			"MINPacket": nil,
		}
	}
	return logrus.Fields{
		"LogicFace": ipd.LogicFace.LogicFaceId,
		"MINPacket": firstIdentifier.ToUri(),
	}
}
