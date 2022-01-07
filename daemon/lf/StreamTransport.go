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
// @Date: 2021/3/22 下午3:24
// @Copyright: MIN-Group；国家重大科技基础设施——未来网络北大实验室；深圳市信息论与未来网络重点实验室
//
package lf

import (
	"errors"
	common2 "minlib/common"
	"minlib/encoding"
	"minlib/packet"
	"net"
)

// StreamTransport
// @Description: 流式通道共用类， TcpTransport, UnixStreamTransport
//			流式传输通信的通用类，主要是提供了一套解决粘包问题的接收方法
//
type StreamTransport struct {
	Transport
	conn    net.Conn
	recvBuf []byte // 数据接收缓冲区
	recvLen uint64 // 当前数据接收缓冲区中的有效数据的长度
}

// Close
// @Description:
// @receiver t
//
func (t *StreamTransport) Close() {
	err := t.conn.Close()
	if err != nil {
		common2.LogWarn(err)
	}
}

// Send
// @Description: 将lpPacket对象编码成字节数组后，通过流式通道发送出去
// @receiver t
// @param lpPacket
//
func (t *StreamTransport) Send(lpPacket *packet.LpPacket) {
	encodeBufLen, encodeBuf := encodeLpPacket2ByteArray(lpPacket)
	if encodeBufLen <= 0 {
		return
	}
	writeLen := 0
	for writeLen < encodeBufLen {
		writeRet, err := t.conn.Write(encodeBuf[:encodeBufLen])
		if err != nil {
			common2.LogError(err, "send to stream transport error:",
				err, ". remote uri: ", t.remoteUri, ", local uri: ", t.localUri)
			t.linkService.logicFace.Shutdown()
			return
		}
		writeLen += writeRet
	}
}

//
// @Description: 从接收缓冲区中读取包并调用linkService的ReceivePacket函数处理包
//			每次收到数据时会调用这个函数从数据接收缓冲区中尝试读出一个LpPacket包。
//			工作流程：
//			（1） 首先，如果数据缓冲区中收到的字节数不足以构成了一个TLV的Type字段，则返回 nil,0 表示还需要等待数据接收
//			（2） 如果解析出来的Type值不等于TlvLpPacket， 表示接收的数据出错了，需要提示调用者关闭Face
//			（3） 如果接收到的数据还小于一个LpPacket的最小长度（不含负载的只含头部的长度）， 表示还需要等待数据接收
//			（4） 如果长度足够，TLV的Length部分，并计算一个LpPacket的总长度 totalPktLen
//			（5） 如果接收到的数据长度小于 totalPktLen ， 则还需要等待后面数据接收
//			（6） 从[]byte中解析出LpPacket，并调用linkService.ReceivePacket(lpPacket) 处理一个完整的LpPacket包
// @receiver t
// @param buf	接收缓冲区分片
// @return error	错误信息
// @return uint64	读取包的长度
//
func (t *StreamTransport) readPktAndDeal(buf []byte, bufLen uint64) (error, uint64) {
	// 如果接收到的数据长度小于 LpPacket type 字段的长度 3字节 则要等待
	if bufLen < uint64(encoding.SizeOfVarNumber(encoding.TlvLpPacket)) {
		return nil, 0
	}
	pktType, err := encoding.ReadVarNumber(buf, 0)
	if err != nil {
		common2.LogWarn(err)
		return err, 0
	}
	// 如果数据类型的 TLV 和 type值不等于   encoding.TlvLpPacket， 则接收出错，应该关闭当前logicFace
	if pktType != encoding.TlvLpPacket {
		common2.LogWarn("receive lpPacket from stream transport type error",
			err, ". remote uri: ", t.remoteUri, ", local uri: ", t.localUri)
		return errors.New("receive lpPacket from stream transport type error"), 0
	}
	// 如果接收到的数据长度小于 LpPacket 的小于长度 则要等待
	if bufLen < uint64(t.linkService.lpPacketHeadSize) {
		return nil, 0
	}
	pktTypeLen := encoding.SizeOfVarNumber(pktType)
	pktLen, err := encoding.ReadVarNumber(buf, encoding.VlInt(pktTypeLen))
	totalPktLen := uint64(pktTypeLen) + uint64(encoding.SizeOfVarNumber(pktLen)) + uint64(pktLen)
	if bufLen >= totalPktLen {
		lpPacket, err := parseByteArray2LpPacket(buf[:totalPktLen])
		if err != nil {
			common2.LogWarn("parse lpPacket error")
		} else {
			t.linkService.ReceivePacket(lpPacket)
		}
		return nil, totalPktLen
	}
	return nil, 0
}

//
// @Description: 接收到数据后，处理包
//			（1） 调用 readPktAndDeal ，传入当前接收的到数据的[]byte，以及接收到的数据长度
//			（2） 如果readPktAndDeal 返回错误，则将错误抛给上层调用者
//			（3） 如果readPktAndDeal没返回错误，且返回的已经被处理的LpPacket长度pktLen大于0, 则循环做以下操作
//					a） 统计已经处理的数据长度dealLen（等于每次处理包长度的总和），如果已经处理的长度小接收数据长度t.recvLen，
//					而且 readPktAndDeal返回的错误为nil，且返回的pktLen > 0 ， 再次调用readPktAndDeal去处理数据
//					b） 如果循环中readPktAndDeal返回的错误不为nil，则终止循环，并将错误报给调用者
//			（4） 如果统计到的总处理长度 dealLen 大小0, 则将已经处理的数据从数据接收缓冲区中删除。删除的方法是将recvBuf[dealLen:t.recvLen]
//				移到 t.recvBuf[:] ， 即将未处理的数据移到接收缓冲区开关，并将接收数据长度t.recvLen 送去 已经处理的长度 dealLen
// @receiver t
// @return error	如果处理包出错，则返回错误信息
//
func (t *StreamTransport) onReceive() error {
	err, pktLen := t.readPktAndDeal(t.recvBuf[:t.recvLen], t.recvLen)
	if err != nil {
		return err
	}
	var dealLen = pktLen
	// 循环多次尝试从接收缓冲区中读出包并处理
	for err == nil && pktLen > 0 && dealLen < t.recvLen {
		err, pktLen = t.readPktAndDeal(t.recvBuf[dealLen:t.recvLen], t.recvLen-dealLen)
		dealLen += pktLen
	}
	if err != nil {
		return err
	}
	if dealLen > 0 {
		copy(t.recvBuf[:], t.recvBuf[dealLen:t.recvLen])
		t.recvLen -= dealLen
	}
	return nil
}

// Receive
// @Description:  用协程调用，不断地从流式通道中读出数据
//			（1） 从流式通道中读出数据，如果读出错，则关闭face
//			（2） 如果读到数据，则调用onReceive尝试处理接收到的数据
//			（3） 如果数据处理出错， 则关闭face
// @receiver t
//
func (t *StreamTransport) Receive() {
	for true {
		recvRet, err := t.conn.Read(t.recvBuf[t.recvLen:])
		if err != nil {
			common2.LogError("recv from stream transport error,the err is:",
				err, ". remote uri: ", t.remoteUri, ", local uri: ", t.localUri)
			t.linkService.logicFace.Shutdown()
			break
		}
		t.recvLen += uint64(recvRet)
		err = t.onReceive()
		if err != nil {
			common2.LogError("recv from stream transport error: ", err, ". remote uri: ", t.remoteUri,
				", local uri: ", t.localUri)
			t.linkService.logicFace.Shutdown()
			break
		}
	}
}
