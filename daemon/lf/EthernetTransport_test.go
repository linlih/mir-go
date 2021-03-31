//
// @Author: weiguohua
// @Description:
// @Version: 1.0.0
// @Date: 2021/3/31 上午11:15
// @Copyright: MIN-Group；国家重大科技基础设施——未来网络北大实验室；深圳市信息论与未来网络重点实验室
//
package lf

import (
	"fmt"
	"minlib/component"
	"minlib/packet"
	"net"
	"testing"
)

func TestEthernetTransport_Send(t *testing.T) {
	//os.Exit(1)
	var LfTb LogicFaceTable
	LfTb.Init()
	var Fsystem LogicFaceSystem
	Fsystem.Init(&LfTb)
	Fsystem.Start()

	//time.Sleep(time.Second * 3)

	localAddr, _ := net.ParseMAC("00:0c:29:fa:de:18")
	remoteAddr, _ := net.ParseMAC("ff:ff:ff:ff:ff:ff")

	interest := new(packet.Interest)
	token := make([]byte, 7000)
	interest.Payload.SetValue(token)
	identifer, _ := component.CreateIdentifierByString("/pkusz")
	interest.SetName(identifer)
	interest.SetTtl(5)

	interest.InterestLifeTime.SetInterestLifeTime(4000)
	fmt.Println("-----------------------------")
	face, faceid := createEtherLogicFace("enp3s0", localAddr, remoteAddr, 8000)
	fmt.Println(faceid)
	fmt.Println(face.GetRemoteUri())
	fmt.Println(face.GetLocalUri())
	fmt.Println(face.transport.GetRemoteAddr())
	fmt.Println(face.transport.GetLocalAddr())
	face.SendInterest(interest)
}
