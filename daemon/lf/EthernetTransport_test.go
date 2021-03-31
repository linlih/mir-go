package lf

import (
	"fmt"
	"minlib/component"
	"minlib/packet"
	"net"
	"testing"
)

func TestEthernetTransport_Send(t *testing.T) {
	var LfTb LogicFaceTable
	LfTb.Init()
	var Fsystem LogicFaceSystem
	Fsystem.Init(&LfTb)
	Fsystem.Start()
	str := "00:0c:29:fa:de:18"
	remote := "ff:ff:ff:ff:ff:ff"
	interest := new(packet.Interest)
	token := make([]byte, 7000)
	interest.Payload.SetValue(token)
	identifer, _ := component.CreateIdentifierByString("/pkusz")
	interest.SetName(identifer)
	interest.SetTtl(5)

	interest.InterestLifeTime.SetInterestLifeTime(4000)
	face, faceid := createEtherLogicFace("ens33", net.HardwareAddr(str), net.HardwareAddr(remote), 8000)
	fmt.Println(faceid)
	face.SendInterest(interest)
}
