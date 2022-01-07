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
// @Author: Lihong Lin
// @Description:
// @Version: 1.0.0
// @Date: 2021/3/31 上午11:15
// @Copyright: MIN-Group；国家重大科技基础设施——未来网络北大实验室；深圳市信息论与未来网络重点实验室
//
package lf_test

import (
	"minlib/packet"
	"mir-go/daemon/common"
	"mir-go/daemon/fw"
	"mir-go/daemon/lf"
	"mir-go/daemon/utils"
	"testing"
)

func TestUnixStreamTransport_Send(t *testing.T) {
	var faceSystem lf.LogicFaceSystem
	var packetValidator fw.PacketValidator
	blockQueue := utils.NewBlockQueue(10)
	packetValidator.Init(1, false, blockQueue)
	var mir common.MIRConfig
	mir.Init()
	// 本地测试，需要在启动faceSystem之前需要关闭TCP/UDP/Unix的收包监听
	mir.SupportTCP = false
	mir.SupportUDP = false
	mir.SupportUnix = false
	faceSystem.Init(&packetValidator, &mir)
	faceSystem.Start()

	logicFace, err := lf.CreateUnixLogicFace("/tmp/mir.sock")
	if err != nil {
		t.Fatal("Create UDP logic face failed", err.Error())
	}

	var interest packet.Interest
	interest.SetNameByString("/min/pkusz")
	interest.SetCanBePrefix(true)
	interest.SetNonce(1234)
	var buf []byte = []byte("hello world!")

	interest.Payload.SetValue(buf[:])
	for i := 0; i < 10; i++ {
		logicFace.SendInterest(&interest)
	}
}
