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
// @Date: 2021/3/16 上午11:35
// @Copyright: MIN-Group；国家重大科技基础设施——未来网络北大实验室；深圳市信息论与未来网络重点实验室
//
package lf

import (
	common2 "minlib/common"
	"minlib/packet"
	"net"
)

// UdpTransport
// @Description:  Udp通信隧道
//
type UdpTransport struct {
	Transport
	conn          *net.UDPConn // UDP句柄
	remoteUdpAddr *net.UDPAddr // 对端UDP地址，用于发送UDP包
}

// Init
// @Description: 	用于接收数据的transport的初始化函数
// @receiver u
// @param conn		UDP句俩
// @param remoteUdpAddr		对端udp地址
//
func (u *UdpTransport) Init(conn *net.UDPConn, remoteUdpAddr *net.UDPAddr) {
	u.conn = conn
	u.localAddr = conn.LocalAddr().String()
	u.localUri = "udp://" + u.localAddr
	if remoteUdpAddr == nil {
		u.remoteAddr = "nil"
		u.remoteUri = "udp://nil"
	} else {
		u.remoteAddr = remoteUdpAddr.String()
		u.remoteUri = "udp://" + u.remoteAddr
		u.remoteUdpAddr = remoteUdpAddr
	}
}

// Close
// @Description: 关闭函数
// @receiver u
//
func (u *UdpTransport) Close() {
	err := u.conn.Close()
	if err != nil {
		common2.LogWarn(err)
	}
}

// Send
// @Description: 往remoteUdpAddr这个地址发送一个UDP包，UDP包里装着一个lpPacket
// @receiver u
// @param lpPacket
//
func (u *UdpTransport) Send(lpPacket *packet.LpPacket) {
	encodeBufLen, encodeBuf := encodeLpPacket2ByteArray(lpPacket)
	if encodeBufLen <= 0 {
		return
	}
	common2.LogTrace("udp send : ", u.remoteUdpAddr.String())
	//_, err := u.conn.Write(encodeBuf)
	_, err := u.conn.WriteToUDP(encodeBuf, u.remoteUdpAddr)
	if err != nil {
		common2.LogWarn(err)
	}
}

//
// @Description: 从UDP句柄中接收到UDP包，并处理
// @receiver u
//
func (u *UdpTransport) doReceive() {
	// 目前用不到
}

// Receive
// @Description: 通过创建一个协程来负责调用这个函数
// @receiver u
//
func (u *UdpTransport) Receive() {
	// TODO 目前用不到
}
