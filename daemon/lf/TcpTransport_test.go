package lf_test

import (
	"minlib/component"
	"minlib/packet"
	"mir-go/daemon/common"
	"mir-go/daemon/fw"
	"mir-go/daemon/lf"
	"mir-go/daemon/utils"
	"sync"
	"testing"
)

func TestTcpTransport_Init(t *testing.T) {
	var LfTb lf.LogicFaceTable
	LfTb.Init()
	var faceSystem lf.LogicFaceSystem
	var packetValidator fw.PacketValidator
	blockQueue := utils.CreateBlockQueue(10)
	packetValidator.Init(100, false, blockQueue)
	var mir common.MIRConfig
	mir.Init()
	faceSystem.Init(&packetValidator, &mir)
	faceSystem.Start()

	id, err := lf.CreateTcpLogicFace("192.168.159.129:13899")
	if err != nil {
		t.Fatal("Create TCP logic face failed", err.Error())
	}
	logicFace := LfTb.GetLogicFacePtrById(id)

	name, err := component.CreateIdentifierByString("/min/pkusz")
	if err != nil {
		t.Fatal("Create Identifier failed", err.Error())
		return
	}
	var interest packet.Interest
	interest.SetName(name)
	interest.SetCanBePrefix(true)
	interest.SetNonce(1234)
	var buf []byte = make([]byte, 8000)
	interest.Payload.SetValue(buf[:])
	for i := 0; i < 100; i++ {
		//time.Sleep(time.Duration(2)*time.Millisecond)
		logicFace.SendInterest(&interest)
	}
	// tcpdump command: sudo tcpdump -i ens33 -nn -s0 -vv -X port 9090

}

//func BenchmarkTcpTransport_Init(b *testing.B) {
//	var LfTb LogicFaceTable
//	LfTb.Init()
//	var Fsystem LogicFaceSystem
//	Fsystem.Init(&LfTb)
//	Fsystem.Start()
//
//	id, err := CreateTcpLogicFace("192.168.159.129:13899")
//	if err != nil {
//		fmt.Println("Create TCP logic face failed", err.Error())
//	}
//	logicFace := LfTb.GetLogicFacePtrById(id)
//
//	name, err := component.CreateIdentifierByString("/min/pkusz")
//	if err != nil {
//		fmt.Println("Create Identifier failed", err.Error())
//		return
//	}
//	var interest packet.Interest
//	interest.SetName(name)
//	interest.SetCanBePrefix(true)
//	interest.SetNonce(1234)
//	var buf []byte = make([]byte,8000)
//	interest.Payload.SetValue(buf[:])
//	b.ResetTimer()
//	for i := 0; i < b.N; i++ {
//		logicFace.SendInterest(&interest)
//	}
//	// tcpdump command: sudo tcpdump -i ens33 -nn -s0 -vv -X port 9090
//}

func testCreateTcpLogicFace(interest packet.Interest, wg *sync.WaitGroup) {

}
