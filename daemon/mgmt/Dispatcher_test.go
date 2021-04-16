package mgmt

import (
	"minlib/component"
	"minlib/encoding"
	"minlib/mgmt"
	"minlib/packet"
	"mir-go/daemon/fw"
	"mir-go/daemon/lf"
	"mir-go/daemon/utils"
	"testing"
	"time"
)

func Test(t *testing.T) {

	var Fsystem lf.LogicFaceSystem
	var packetValidator = &fw.PacketValidator{}
	blockQueue := utils.CreateBlockQueue(100)
	packetValidator.Init(100, false, blockQueue)
	Fsystem.Init(packetValidator, nil)
	Fsystem.Start()

	fibManager := CreateFibManager()
	faceManager := CreateFaceManager()
	csManager := CreateCsManager()
	dispatcher := CreateDispatcher(nil)
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
	identifier, _ = component.CreateIdentifierByString("/min-mir/mgmt/localhop/face-mgmt/list")
	interest.SetName(identifier)
	face.SendInterest(interest)

	face.SendInterest(interest)

	time.Sleep(time.Minute)
}
