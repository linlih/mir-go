package lf_test

import (
	"flag"
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
	"sync"
	"testing"
	"time"
)

func TestTcpTransport_Init(t *testing.T) {
	var faceSystem lf.LogicFaceSystem
	var packetValidator fw.PacketValidator
	blockQueue := utils.BlockQueue{}
	packetValidator.Init(100, false, &blockQueue)
	var mir common.MIRConfig
	mir.Init()
	faceSystem.Init(&packetValidator, &mir)
	faceSystem.Start()

	logicFace, err := lf.CreateTcpLogicFace("192.168.159.129:13899", 0)
	if err != nil {
		t.Fatal("Create TCP logic face failed", err.Error())
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
	var buf []byte = make([]byte, 1300)
	interest.Payload.SetValue(buf[:])
	counter := 0
	//fmt.Println(faceid)

	//var keychain security.KeyChain
	//keychain.Init()
	//keychain.CreateIdentityByName("/yb","123123123123")
	start := time.Now()
	for {
		logicFace.SendInterest(&interest)
		counter++
		//time.Sleep(30 * time.Microsecond)
		//common2.LogInfo(counter)
		if counter == 1000000 {
			eclipase := time.Since(start)
			common2.LogInfo(eclipase)
			break
		}
	}
	// tcpdump command: sudo tcpdump -i ens33 -nn -s0 -vv -X port 9090

}
func tcpTransportSend(remoteAddr string, payloadSize int, nums int, wg *sync.WaitGroup) {
	logicFace, err := lf.CreateTcpLogicFace(remoteAddr+":13899", 0)
	if err != nil {
		fmt.Println("Create Tcp logic face failed", err.Error())
		return
	}

	var interest packet.Interest
	interest.SetNameByString("/min/pkusz")
	interest.SetCanBePrefix(true)
	interest.SetNonce(1234)

	interest.Payload.SetValue(utils.RandomBytes(payloadSize))

	// tcpdump command: sudo tcpdump -i ens33 -nn -s0 -vv -X port 13899
	start := time.Now()
	for i := 0; i < nums; i++ {
		logicFace.SendInterest(&interest)
	}
	eclipse := time.Since(start)
	common2.LogInfo(eclipse)
	wg.Done()
}

func tcpTransportSendAndSign(remoteAddr string, payloadSize int, nums int, wg *sync.WaitGroup) {
	logicFace, err := lf.CreateTcpLogicFace(remoteAddr+":13899", 0)
	if err != nil {
		fmt.Println("Create Tcp logic face failed", err.Error())
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
	start := time.Now()
	for i := 0; i < nums; i++ {
		logicFace.SendInterest(&interest)
	}
	eclipse := time.Since(start)
	common2.LogInfo(eclipse)
	wg.Done()
}

// 增加命令行参数后的测试命令如下：
// go test . -test.run "TestTcpTransport_SpeedAnd" -v -count=1 -args -remoteAddr=192.168.0.8 -payloadSize=2000 -nums=2 -routineNum=2
var remoteTcpAddr = flag.String("remoteTcpAddr", "127.0.0.1", "Tcp remote connect address")
var TcpNums = flag.Int("TcpNums", 1, "number of Tcp interest packet")
var TcpPayloadSize = flag.Int("TcpPayloadSize", 1300, "payload's size of Tcp sending interest packet")
var TcpRoutineNum = flag.Int("TcpRoutineNum", 1, "number of routine to send Tcp interest")

func TestTcpTransport_Speed(t *testing.T) {
	var faceSystem lf.LogicFaceSystem
	var packetValidator fw.PacketValidator
	blockQueue := utils.BlockQueue{}
	packetValidator.Init(100, true, &blockQueue)
	var mir common.MIRConfig
	mir.Init()
	faceSystem.Init(&packetValidator, &mir)
	faceSystem.Start()
	flag.Parse()

	var goRoutineNum int = *TcpRoutineNum
	var wg sync.WaitGroup
	wg.Add(goRoutineNum)
	for i := 0; i < goRoutineNum; i++ {
		go tcpTransportSend(*remoteTcpAddr, *TcpPayloadSize, *TcpNums, &wg)
	}
	wg.Wait()
}

func TestTcpTransport_SpeedAndSign(t *testing.T) {
	var faceSystem lf.LogicFaceSystem
	var packetValidator fw.PacketValidator
	blockQueue := utils.BlockQueue{}
	packetValidator.Init(100, true, &blockQueue)
	var mir common.MIRConfig
	mir.Init()
	faceSystem.Init(&packetValidator, &mir)
	faceSystem.Start()
	flag.Parse()

	var goRoutineNum int = *TcpRoutineNum
	var wg sync.WaitGroup
	wg.Add(goRoutineNum)
	for i := 0; i < goRoutineNum; i++ {
		go tcpTransportSendAndSign(*remoteTcpAddr, *TcpPayloadSize, *TcpNums, &wg)
	}
	wg.Wait()
}

func TestTcpTransport_Receive(t *testing.T) {
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
		common2.LogInfo("======================")
		for _, face := range faceSystem.LogicFaceTable().GetAllFaceList() {
			common2.LogInfo(face.LogicFaceId, "=>", face.GetCounter())
		}
		common2.LogInfo("======================")
	}

}
