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
// @Date: 2021/3/16 上午11:32
// @Copyright: MIN-Group；国家重大科技基础设施——未来网络北大实验室；深圳市信息论与未来网络重点实验室
//
package lf

import (
	common2 "minlib/common"
	"minlib/encoding"
	"minlib/packet"
)

// Transport
// @Description:  Tranport共用类
//
type Transport struct {
	localAddr   string
	remoteAddr  string
	localUri    string
	remoteUri   string
	linkService *LinkService
}

//
// @Description: 从[]byte中解析出LpPacket
// @receiver e
// @param pkt
// @return *packet.LpPacket	解析出的包
// @return error		解析失败错误
//
func parseByteArray2LpPacket(buf []byte) (*packet.LpPacket, error) {
	block, err := encoding.CreateBlockByBuffer(buf, true)
	if err != nil {
		common2.LogWarn(err)
		return nil, err
	}
	if !block.IsValid() {
		common2.LogWarn("recv packet from face invalid")
		return nil, err
	}
	var lpPacket packet.LpPacket
	err = lpPacket.WireDecode(block)
	if err != nil {
		common2.LogWarn("parse to lpPacket error")
		return nil, err
	}
	return &lpPacket, nil
}

//
// @Description: 	将lpPacket编码成byte数组
// @receiver t
// @param lpPacket
// @return int     编码后byte数组的长度
// @return []byte	编码得到的byte数组
//
func encodeLpPacket2ByteArray(lpPacket *packet.LpPacket) (int, []byte) {
	var encoder encoding.Encoder
	err := encoder.EncoderReset(encoding.MaxPacketSize, 0)
	encodeBufLen, err := lpPacket.WireEncode(&encoder)
	if err != nil {
		common2.LogWarn(err)
		return -1, nil
	}
	encodeBuf, err := encoder.GetBuffer()
	if err != nil {
		common2.LogWarn(err)
		return -1, nil
	}
	return encodeBufLen, encodeBuf
}

// GetRemoteUri
// @Description:
// @receiver t
// @return string
//
func (t *Transport) GetRemoteUri() string {
	return t.remoteUri
}

// GetLocalUri
// @Description:
// @receiver t
// @return string
//
func (t *Transport) GetLocalUri() string {
	return t.localUri
}

// GetRemoteAddr
// @Description:
// @receiver t
// @return string
//
func (t *Transport) GetRemoteAddr() string {
	return t.remoteAddr
}

// GetLocalAddr
// @Description:
// @receiver t
// @return string
//
func (t *Transport) GetLocalAddr() string {
	return t.localAddr
}
