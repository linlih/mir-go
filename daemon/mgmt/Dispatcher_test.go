package mgmt

import (
	"minlib/component"
	"minlib/encoding"
	"minlib/mgmt"
	"minlib/packet"
	"mir-go/daemon/fw"
	"mir-go/daemon/lf"
	"testing"
	"time"
)

func Test(t *testing.T) {
	fibManager := CreateFibManager()
	faceManager := CreateFaceManager()
	csManager := CreateCsManager()
	dispatcher := CreateDispatcher()
	topPrefix, _ := component.CreateIdentifierByString("/min-mir/mgmt/localhost")
	dispatcher.AddTopPrefix(topPrefix)
	fibManager.Init(dispatcher)
	faceManager.Init(dispatcher)
	csManager.Init(dispatcher)
	// FIX:下面这两行暂时保留 后面可能需要删除
	var LfTb lf.LogicFaceTable
	LfTb.Init()
	var Fsystem lf.LogicFaceSystem
	var packetValidator fw.PacketValidator
	blockQueue := fw.BlockQueue{}
	packetValidator.Init(100, false, &blockQueue)
	Fsystem.Init(&LfTb, &packetValidator)
	Fsystem.Start()

	faceServer, faceClient := lf.CreateInnerLogicFacePair()
	dispatcher.FaceClient = faceClient
	topPrefix, _ = component.CreateIdentifierByString("/min-mir/mgmt/localhop")
	dispatcher.AddTopPrefix(topPrefix)
	fibManager.GetFib().AddOrUpdate(topPrefix, faceServer, 0)
	dispatcher.Start()

	identifier, _ := component.CreateIdentifierByString("/min-mir/mgmt/localhop/fib-mgmt/delete")
	logicfaceId := fibManager.fib.FindExactMatch(topPrefix).GetNextHops()[0].LogicFace.LogicFaceId
	face := lf.GLogicFaceTable.GetLogicFacePtrById(logicfaceId)
	face.Start()

	interest := &packet.Interest{}
	parameters := &mgmt.ControlParameters{}
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
	identifier, _ = component.CreateIdentifierByString("/min-mir/mgmt/localhop/fib-mgmt/list")
	interest.SetName(identifier)
	face.SendInterest(interest)

	time.Sleep(time.Minute)
}
