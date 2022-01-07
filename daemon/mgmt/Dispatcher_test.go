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

// Package mgmt
// @Author: yzy
// @Description:
// @Version: 1.0.0
// @Date: 2021/12/21 4:32 PM
// @Copyright: MIN-Group；国家重大科技基础设施——未来网络北大实验室；深圳市信息论与未来网络重点实验室
//

package mgmt

import (
	"minlib/component"
	"minlib/encoding"
	"minlib/packet"
	"minlib/utils"
	"mir-go/daemon/fw"
	"mir-go/daemon/lf"
	"testing"
	"time"
)

func Test(t *testing.T) {

	var Fsystem lf.LogicFaceSystem
	var packetValidator = &fw.PacketValidator{}
	blockQueue := utils.NewBlockQueue(100)
	packetValidator.Init(100, false, blockQueue)
	Fsystem.Init(packetValidator, nil)
	Fsystem.Start()

	fibManager := CreateFibManager()
	faceManager := CreateFaceManager()
	csManager := CreateCsManager()
	dispatcher := CreateDispatcher(nil, nil)
	topPrefix, _ := component.CreateIdentifierByString("/min-mir/mgmt/localhost")
	dispatcher.AddTopPrefix(topPrefix, nil, nil)
	fibManager.Init(dispatcher, Fsystem.LogicFaceTable())
	faceManager.Init(dispatcher, Fsystem.LogicFaceTable())
	csManager.Init(dispatcher, Fsystem.LogicFaceTable())

	faceServer, faceClient := lf.CreateInnerLogicFacePair()
	dispatcher.FaceClient = faceClient
	topPrefix, _ = component.CreateIdentifierByString("/min-mir/mgmt/localhop")
	dispatcher.AddTopPrefix(topPrefix, nil, nil)
	fibManager.GetFib().AddOrUpdate(topPrefix, faceServer, 0)
	dispatcher.Start()

	identifier, _ := component.CreateIdentifierByString("/min-mir/mgmt/localhop/fib-mgmt/delete")
	logicfaceId := fibManager.fib.FindExactMatch(topPrefix).GetNextHops()[0].LogicFace.LogicFaceId
	face := fibManager.logicFaceTable.GetLogicFacePtrById(logicfaceId)
	face.Start()

	interest := &packet.Interest{}
	parameters := &component.ControlParameters{}
	prefix, _ := component.CreateIdentifierByString("/min")
	parameters.SetPrefix(prefix)
	parameters.SetCost(10)
	parameters.SetLogicFaceId(0)
	var encoder = &encoding.Encoder{}
	encoder.EncoderReset(encoding.MaxPacketSize, 0)
	parameters.WireEncode(encoder)
	buf, _ := encoder.GetBuffer()
	identifier.Append(component.CreateIdentifierComponentByByteArray(buf))
	interest.SetName(identifier)
	face.SendInterest(interest)
	identifier, _ = component.CreateIdentifierByString("/min-mir/mgmt/localhop/face-mgmt/list")
	interest.SetName(identifier)
	face.SendInterest(interest)

	face.SendInterest(interest)

	time.Sleep(time.Minute)
}
