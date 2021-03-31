<<<<<<< HEAD
=======
//
// @Author: weiguohua
// @Description: 
// @Version: 1.0.0
// @Date: 2021/3/31 上午11:15  
// @Copyright: MIN-Group；国家重大科技基础设施——未来网络北大实验室；深圳市信息论与未来网络重点实验室
//
>>>>>>> master
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
<<<<<<< HEAD
	str := "00:0c:29:fa:de:18"
	remote := "ff:ff:ff:ff:ff:ff"
=======

	str := "00:0c:29:fa:de:18"
	remote := "ff:ff:ff:ff:ff:ff"

>>>>>>> master
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
