//
// @Author: weiguohua
// @Description:
// @Version: 1.0.0
// @Date: 2021/3/31 上午11:15
// @Copyright: MIN-Group；国家重大科技基础设施——未来网络北大实验室；深圳市信息论与未来网络重点实验室
//

package lf_test

import (
	"fmt"
	"minlib/component"
	"minlib/packet"
	"mir-go/daemon/fw"
	"mir-go/daemon/lf"
	"net"
	"testing"
)

func TestEthernetTransport_Send(t *testing.T) {
	var LfTb lf.LogicFaceTable
	LfTb.Init()
	var Fsystem lf.LogicFaceSystem
	var packetValidator fw.PacketValidator
	blockQueue := fw.BlockQueue{}
	packetValidator.Init(100, false, &blockQueue)
	Fsystem.Init(&LfTb, &packetValidator)
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
	//localMac, err := net.ParseMAC(str)
	_, err := net.ParseMAC(str)
	if err != nil {
		fmt.Println("local mac", err)
	}
	//remoteMac, err1 := net.ParseMAC(remote)
	_, err1 := net.ParseMAC(remote)
	if err1 != nil {
		fmt.Println("local mac", err1)
	}
	//face, faceid := createEtherLogicFace("ens33", localMac, remoteMac, 1500)
	//fmt.Println(faceid)
	//face.SendInterest(interest)
}
