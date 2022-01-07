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
// @Date: 2021/3/31 上午11:15
// @Copyright: MIN-Group；国家重大科技基础设施——未来网络北大实验室；深圳市信息论与未来网络重点实验室
//

package lf

import (
	common2 "minlib/common"
	"minlib/packet"
)

// InnerTransport
// @Description:  用于MIR内部模块间通信的，使用 chan 进行通信， 使用一个用于发数据的chan和
type InnerTransport struct {
	Transport
	sendChan chan<- *packet.LpPacket
	recvChan <-chan *packet.LpPacket
}

func (i *InnerTransport) Init(sendChan chan<- *packet.LpPacket, recvChan <-chan *packet.LpPacket) {
	i.sendChan = sendChan
	i.recvChan = recvChan
	i.localAddr = "inner://nil"
	i.localUri = "inner://nil"
	i.remoteAddr = "inner://nil"
	i.remoteUri = "inner://nil"
}

// Close
// @Description: 只能关闭发数据的chan
// @receiver i
//
func (i *InnerTransport) Close() {
	close(i.sendChan)
}

// Send
// @Description: 往chan里发lpPacket
// @receiver i
// @param lpPacket
//
func (i *InnerTransport) Send(lpPacket *packet.LpPacket) {
	i.sendChan <- lpPacket
}

// Receive
// @Description: 从chan中接收lpPacket
// @receiver i
//
func (i *InnerTransport) Receive() {
	for true {
		lpPacket, ok := <-i.recvChan
		if !ok {
			common2.LogInfo("inner channel has been closed")
			break
		}
		i.linkService.ReceivePacket(lpPacket)
	}
}
