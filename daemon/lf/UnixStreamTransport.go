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

import "net"

type UnixStreamTransport struct {
	StreamTransport
}

func (u *UnixStreamTransport) Init(conn net.Conn) {
	u.conn = conn
	u.localAddr = conn.LocalAddr().String()
	u.localUri = "unix://" + u.localAddr
	u.remoteAddr = conn.RemoteAddr().String()
	u.remoteUri = "unix://" + u.remoteAddr
	u.recvBuf = make([]byte, 1024*1024*4)
	u.recvLen = 0
}
