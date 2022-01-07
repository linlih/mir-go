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
// @Author: weiguohua
// @Description:
// @Version: 1.0.0
// @Date: 2021/3/17 下午6:03
// @Copyright: MIN-Group；国家重大科技基础设施——未来网络北大实验室；深圳市信息论与未来网络重点实验室
//
package lf

import "minlib/packet"

// ITransport
// @Description:  Tranport 接口， 便于LogicFace声明成员。logicFace模块中的每一种tranport都必须实现ITransport声明的方法
//
type ITransport interface {
	// Close
	// @Description:  关闭
	//
	Close()
	// Send
	// @Description: 发送一个lpPacket
	// @param lpPacket
	//
	Send(lpPacket *packet.LpPacket)
	// Receive
	// @Description: 从网络中接收到一段数据
	//
	Receive()
	// GetRemoteUri
	// @Description: 获得Transport的对端地址
	//			格式 ：
	//			TCP  tcp://192.238.3.3:7890
	//			UDP  udp://192.238.3.3:7890
	//			ether  ether://fc:aa:14:cf:a6:97
	//			unix  unix:///tmp/mirsock
	// @return string	对端地址
	//
	GetRemoteUri() string
	// GetLocalUri
	// @Description: 获得Transport的本机地址
	//			格式 ：
	//			TCP  tcp://192.238.3.3:7890
	//			UDP  udp://192.238.3.3:7890
	//			ether  ether://fc:aa:14:cf:a6:97
	//			unix  unix:///tmp/mirsock
	// @return string	本机地址
	//
	GetLocalUri() string
	// GetRemoteAddr @Description: 获得Transport的对端地址
	//			格式 ：
	//			TCP  192.238.3.3:7890
	//			UDP  192.238.3.3:7890
	//			ether  fc:aa:14:cf:a6:97
	//			unix  /tmp/mirsock
	// @return string	对端地址
	//
	GetRemoteAddr() string
	// GetLocalAddr
	// @Description: 获得Transport的本机地址
	//			格式 ：
	//			TCP  192.238.3.3:7890
	//			UDP  192.238.3.3:7890
	//			ether  fc:aa:14:cf:a6:97
	//			unix  /tmp/mirsock
	// @return string	本机地址
	//
	GetLocalAddr() string
}
