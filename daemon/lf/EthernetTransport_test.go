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

package lf_test

import (
	"flag"
	"fmt"
	"math/rand"
	common2 "minlib/common"
	"minlib/component"
	"minlib/packet"
	"minlib/security"
	"minlib/utils"
	"mir-go/daemon/common"
	"mir-go/daemon/fw"
	"mir-go/daemon/lf"
	"net"
	"net/http"
	_ "net/http/pprof"
	"sync"
	"testing"
	"time"
)

// 增加命令行参数后的测试命令如下：
// go test . -test.run "TestUdpTransport_SpeedAnd" -v -count=1 -args -remoteAddr=192.168.0.8 -payloadSize=2000 -nums=2 -routineNum=2
var remoteMacAddr = flag.String("remoteMacAddr", "00:0c:29:a1:35:bf", "Ethernet remote connect address")
var macNums = flag.Int("nums", 1, "number of UDP interest packet")
var macPayloadSize = flag.Int("payloadSize", 1300, "payload's size of UDP sending interest packet")
var macRoutineNum = flag.Int("routineNum", 1, "number of routine to send UDP interest")

func TestEthernetTransport_Send(t *testing.T) {
	var faceSystem lf.LogicFaceSystem
	var packetValidator fw.PacketValidator
	blockQueue := utils.NewBlockQueue(10)
	packetValidator.Init(100, false, blockQueue)
	var mir common.MIRConfig
	mir.Init()
	faceSystem.Init(&packetValidator, &mir)
	faceSystem.Start()
	time.Sleep(1 * time.Second)

	flag.Parse()
	interest := createInterest()
	//localMac, err := net.ParseMAC(str)
	remoteAddr, _ := net.ParseMAC(*remoteMacAddr)
	logicFace, faceErr := lf.CreateEtherLogicFace("ens33", remoteAddr)
	if faceErr != nil {
		common2.LogError(faceErr)
	}
	//指定pprof对外提供的http服务的ip和端口，配置为0.0.0.0表示可以非本机访问
	go func() {
		http.ListenAndServe("0.0.0.0:9999", nil)
	}()
	//counter:=0
	//start := time.Now()
	//for {
	//	logicFace.SendInterest(interest)
	//	counter++
	//	if counter == 1000000 {
	//		eclipase := time.Since(start)
	//		common2.LogInfo(eclipase)
	//		break
	//	}
	//}
	var wg sync.WaitGroup
	wg.Add(*macRoutineNum)
	for i := 0; i < *macRoutineNum; i++ {
		go EtherTransportSend2(interest, logicFace, &wg)
	}
	wg.Wait()
}
func EtherTransportSend2(interest *packet.Interest, logicFace *lf.LogicFace, wg *sync.WaitGroup) {
	counter := 0
	start := time.Now()
	for {
		logicFace.SendInterest(interest)
		//time.Sleep(1*time.Nanosecond)
		counter++
		//time.Sleep(30 * time.Microsecond)
		//common2.LogInfo(counter)
		if counter == 1000000 {
			eclipase := time.Since(start)
			common2.LogInfo(eclipase)
			break
		}
	}
	wg.Done()
}
func createInterest() *packet.Interest {
	interest := new(packet.Interest)
	token := randByte()
	interest.Payload.SetValue(token)
	identifier, _ := component.CreateIdentifierByString("/pkusz")
	interest.SetName(identifier)
	interest.SetTtl(5)
	interest.InterestLifeTime.SetInterestLifeTime(4000)

	return interest
}

func randByte() []byte {
	token := make([]byte, 8000)
	rand.Read(token)

	return token
}

//benchmark test
func EtherTransportSend(faceSystem *lf.LogicFaceSystem, payloadSize int, volume int, wg *sync.WaitGroup) {
	remote := "00:0c:29:a1:35:bf"
	remoteAddr, _ := net.ParseMAC(remote)
	logicFace, err := lf.CreateEtherLogicFace("ens33", remoteAddr)
	if err != nil {
		fmt.Println("Create Ethernet logic face failed", err.Error())
		return
	}

	var interest packet.Interest
	interest.SetNameByString("/pkusz")
	interest.SetCanBePrefix(true)
	interest.SetNonce(1234)

	interest.Payload.SetValue(utils.RandomBytes(payloadSize))

	// tcpdump command: sudo tcpdump -i ens33 -nn -s0 -vv -X port 13899
	for i := 0; i < volume; i++ {
		logicFace.SendInterest(&interest)
	}
	wg.Done()
}

func TestEtherTransport_Speed(t *testing.T) {
	var faceSystem lf.LogicFaceSystem
	var packetValidator fw.PacketValidator
	blockQueue := utils.BlockQueue{}
	packetValidator.Init(100, true, &blockQueue)
	var mir common.MIRConfig
	mir.Init()
	faceSystem.Init(&packetValidator, &mir)
	faceSystem.Start()

	var goRoutineNum int = 3
	var wg sync.WaitGroup
	wg.Add(goRoutineNum)
	for i := 0; i < goRoutineNum; i++ {
		go EtherTransportSend(&faceSystem, 8000, 1000000, &wg)
	}
	wg.Wait()
}

func EtherTransportSendAndSign(faceSystem *lf.LogicFaceSystem, payloadSize int, volume int, wg *sync.WaitGroup) {
	logicFace, err := lf.CreateUdpLogicFace("192.168.0.9:13899")
	if err != nil {
		fmt.Println("Create UDP logic face failed", err.Error())
		return
	}
	var interest packet.Interest
	interest.SetNameByString("/min/pkusz")
	interest.SetCanBePrefix(true)
	interest.SetNonce(1234)
	interest.Payload.SetValue(utils.RandomBytes(payloadSize))

	keyChain, err := security.CreateKeyChain()
	if err != nil {
		fmt.Println("Create KeyChain failed ", err.Error())
		return
	}
	// 测试前需要保证两条机器有相同的秘钥，需要用程序在一台主机上生成下，再拷贝到另外一台机器上
	i := keyChain.IdentityManager.GetIdentityByName("/pkusz")
	keyChain.SetCurrentIdentity(i, "pkusz123pkusz123")
	keyChain.SignInterest(&interest)

	// tcpdump command: sudo tcpdump -i ens33 -nn -s0 -vv -X port 13899
	for i := 0; i < volume; i++ {
		logicFace.SendInterest(&interest)
	}
	wg.Done()
}
func TestEtherTransport_SpeedAndSign(t *testing.T) {
	var faceSystem lf.LogicFaceSystem
	var packetValidator fw.PacketValidator
	blockQueue := utils.BlockQueue{}
	packetValidator.Init(100, true, &blockQueue)
	var mir common.MIRConfig
	mir.Init()
	faceSystem.Init(&packetValidator, &mir)
	faceSystem.Start()

	var goRoutineNum int = 3
	var wg sync.WaitGroup
	wg.Add(goRoutineNum)
	for i := 0; i < goRoutineNum; i++ {
		go EtherTransportSendAndSign(&faceSystem, 1500, 10000, &wg)
	}
	wg.Wait()
}
