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
// @Date: 2021/3/16 上午11:37
// @Copyright: MIN-Group；国家重大科技基础设施——未来网络北大实验室；深圳市信息论与未来网络重点实验室
//
package lf

import (
	common2 "minlib/common"
	"mir-go/daemon/common"
	"mir-go/daemon/utils"
	"net"
	"strconv"
)

// TcpListener
// @Description:  TCP端口监听器，用于接收远程mir的TCP连接请求，为新连接创建
//			并启动一个TCP-Transport类型的LogicFace
//
type TcpListener struct {
	TcpPort  uint16       // TCP端口号
	listener net.Listener // TCP监听句柄
	config   *common.MIRConfig
}

// Init
// @Description: 	初始化TCP监听器
// @receiver t
// @param logicFaceTable  全局logicFace表指针
//
func (t *TcpListener) Init(config *common.MIRConfig) {
	t.TcpPort = uint16(config.TCPPort)
	t.config = config
}

//
// @Description: 创建一个TCP类型的logicFace
// @receiver t
// @param conn	新TCP连接句柄
//
func (t *TcpListener) tryCreateTcpLogicFace(conn net.Conn) {
	createTcpLogicFace(conn, 0)
}

//
// @Description: 接收TCP连接，并创建TCP类型的LogicFace
// @receiver t
//
func (t *TcpListener) accept() {
	for true {
		newConnect, err := t.listener.Accept()
		if err != nil {
			common2.LogFatal(err)
		}
		t.tryCreateTcpLogicFace(newConnect)
	}
}

// Start
// @Description:  启动监听协程
// @receiver t
//
func (t *TcpListener) Start() {
	listener, err := net.Listen("tcp", "0.0.0.0:"+strconv.Itoa(int(t.TcpPort)))
	if err != nil {
		common2.LogFatal(err)
		return
	}
	t.listener = listener
	utils.GoroutineNoPanic(t.accept)
}
