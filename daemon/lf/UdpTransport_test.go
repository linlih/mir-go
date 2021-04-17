//
// @Author: Lihong Lin
// @Description:
// @Version: 1.0.0
// @Date: 2021/3/31 下午7:45
// @Copyright: MIN-Group；国家重大科技基础设施——未来网络北大实验室；深圳市信息论与未来网络重点实验室
//
package lf_test

import (
	"fmt"
	common2 "minlib/common"
	"minlib/component"
	"minlib/packet"
	"minlib/security"
	"mir-go/daemon/common"
	"mir-go/daemon/fw"
	"mir-go/daemon/lf"
	"mir-go/daemon/utils"
	"net/http"
	_ "net/http/pprof"
	"sync"
	"testing"
	"time"
)

/**************************************************************
 * 测试场景说明：
 * 两台虚拟机，VM1和VM2
 * IP信息： VM1 192.168.0.9,  VM2 192.168.0.8
 * VM1 先执行 TestUdpTransport_Receive 函数进行13899端口的UDP收包监听
 * VM2 后执行 TestUdpTransport_Init    函数向VM1的13899发送一个兴趣包
 * VM1 会在终端中打印 "recv from :" 等信息，并正确解码相应的兴趣包
 ***************************************************************/
func TestUdpTransport_Init(t *testing.T) {
	var faceSystem lf.LogicFaceSystem
	var packetValidator fw.PacketValidator
	blockQueue := utils.BlockQueue{}
	packetValidator.Init(100, true, &blockQueue)
	var mir common.MIRConfig
	mir.Init()
	faceSystem.Init(&packetValidator, &mir)
	faceSystem.Start()

	id, err := lf.CreateUdpLogicFace("192.168.0.9:13899")
	if err != nil {
		t.Fatal("Create UDP logic face failed", err.Error())
	}

	name, err := component.CreateIdentifierByString("/min/pkusz")
	if err != nil {
		t.Fatal("Create Identifier failed", err.Error())
		return
	}
	var interest packet.Interest
	interest.SetName(name)
	interest.SetCanBePrefix(true)
	interest.SetNonce(1234)
	var buf []byte = []byte("hello world!")

	interest.Payload.SetValue(buf[:])
	logicFace := faceSystem.LogicFaceTable().GetLogicFacePtrById(id)

	// tcpdump command: sudo tcpdump -i ens33 -nn -s0 -vv -X port 13899
	logicFace.SendInterest(&interest)
}

func udpTransportSend(faceSystem *lf.LogicFaceSystem, payloadSize int, volume int, wg *sync.WaitGroup) {
	id, err := lf.CreateUdpLogicFace("192.168.0.9:13899")
	if err != nil {
		fmt.Println("Create UDP logic face failed", err.Error())
		return
	}

	var interest packet.Interest
	interest.SetNameByString("/min/pkusz")
	interest.SetCanBePrefix(true)
	interest.SetNonce(1234)

	interest.Payload.SetValue(utils.RandomBytes(payloadSize))
	logicFace := faceSystem.LogicFaceTable().GetLogicFacePtrById(id)

	// tcpdump command: sudo tcpdump -i ens33 -nn -s0 -vv -X port 13899
	for i := 0; i < volume; i++ {
		logicFace.SendInterest(&interest)
	}
	wg.Done()
}

func udpTransportSendAndSign(faceSystem *lf.LogicFaceSystem, payloadSize int, volume int, wg *sync.WaitGroup) {
	id, err := lf.CreateUdpLogicFace("192.168.0.9:13899")
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
	i := keyChain.IdentifyManager.GetIdentifyByName("/pkusz")
	keyChain.SetCurrentIdentity(i, "pkusz123pkusz123")
	keyChain.SignInterest(&interest)

	logicFace := faceSystem.LogicFaceTable().GetLogicFacePtrById(id)
	// tcpdump command: sudo tcpdump -i ens33 -nn -s0 -vv -X port 13899
	for i := 0; i < volume; i++ {
		logicFace.SendInterest(&interest)
	}
	wg.Done()
}

func TestUdpTransport_Speed(t *testing.T) {
	var faceSystem lf.LogicFaceSystem
	var packetValidator fw.PacketValidator
	blockQueue := utils.BlockQueue{}
	packetValidator.Init(100, true, &blockQueue)
	var mir common.MIRConfig
	mir.Init()
	faceSystem.Init(&packetValidator, &mir)
	faceSystem.Start()

	var goRoutineNum int = 1
	var wg sync.WaitGroup
	wg.Add(goRoutineNum)
	for i := 0; i < goRoutineNum; i++ {
		go udpTransportSend(&faceSystem, 8000, 1000000, &wg)
	}
	wg.Wait()
}

func TestUdpTransport_SpeedAndSign(t *testing.T) {
	var faceSystem lf.LogicFaceSystem
	var packetValidator fw.PacketValidator
	blockQueue := utils.BlockQueue{}
	packetValidator.Init(100, true, &blockQueue)
	var mir common.MIRConfig
	mir.Init()
	faceSystem.Init(&packetValidator, &mir)
	faceSystem.Start()

	var goRoutineNum int = 1
	var wg sync.WaitGroup
	wg.Add(goRoutineNum)
	for i := 0; i < goRoutineNum; i++ {
		go udpTransportSendAndSign(&faceSystem, 1500, 10000, &wg)
	}
	wg.Wait()
}

func TestUdpTransport_Receive(t *testing.T) {
	var faceSystem lf.LogicFaceSystem
	var packetValidator fw.PacketValidator
	blockQueue := utils.CreateBlockQueue(10)
	packetValidator.Init(1, false, blockQueue)
	var mir common.MIRConfig
	mir.Init()
	faceSystem.Init(&packetValidator, &mir)
	faceSystem.Start()

	go func() {
		http.ListenAndServe("0.0.0.0:9999", nil)
	}()

	for true {
		//time.Sleep(10 * time.Second)
		//fmt.Println("等待收包")
		time.Sleep(3 * time.Second)
		common2.LogInfo("\n\n======")
		for _, face := range faceSystem.LogicFaceTable().GetAllFaceList() {
			common2.LogInfo(face.LogicFaceId, "=>", face.GetCounter())
		}
		common2.LogInfo("======\n\n")
	}

}
