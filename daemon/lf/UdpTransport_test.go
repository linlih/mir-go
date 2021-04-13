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
	"minlib/component"
	"minlib/packet"
	"mir-go/daemon/common"
	"mir-go/daemon/fw"
	"mir-go/daemon/lf"
	"mir-go/daemon/utils"
	"net/http"
	_ "net/http/pprof"
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
	var LfTb lf.LogicFaceTable
	LfTb.Init()
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
	var buf []byte = []byte("hello world!")
	interest.Payload.SetValue(buf[:])

	// tcpdump command: sudo tcpdump -i ens33 -nn -s0 -vv -X port 13899
	logicFace.SendInterest(&interest)
}

func TestUdpTransport_Receive(t *testing.T) {
	var LfTb lf.LogicFaceTable
	LfTb.Init()
	var faceSystem lf.LogicFaceSystem
	var packetValidator fw.PacketValidator
	blockQueue := utils.CreateBlockQueue(10)
	packetValidator.Init(2, false, blockQueue)
	var mir common.MIRConfig
	mir.Init()
	faceSystem.Init(&packetValidator, &mir)
	faceSystem.Start()

	go func() {
		http.ListenAndServe("0.0.0.0:9999", nil)
	}()

	for true {
		time.Sleep(10 * time.Second)
		fmt.Println("等待收包")
	}
}
