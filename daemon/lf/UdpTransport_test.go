//
// @Author: Lihong Lin
// @Description:
// @Version: 1.0.0
// @Date: 2021/3/31 下午7:45 
// @Copyright: MIN-Group；国家重大科技基础设施——未来网络北大实验室；深圳市信息论与未来网络重点实验室
//
package lf

import (
	"minlib/component"
	"minlib/packet"
	"testing"
)

func TestUdpTransport_Init(t *testing.T) {
	var LfTb LogicFaceTable
	LfTb.Init()
	var Fsystem LogicFaceSystem
	Fsystem.Init(&LfTb)
	Fsystem.Start()

	id, err := CreateUdpLogicFace("192.168.0.2:9090")
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

	// tcpdump command: sudo tcpdump -i ens33 -nn -s0 -vv -X port 9090
	logicFace.SendInterest(&interest)
}