//
// @Author: Lihong Lin
// @Description:
// @Version: 1.0.0
// @Date: 2021/4/16 下午8:45
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
	blockQueue := utils.CreateBlockQueue(10)
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
