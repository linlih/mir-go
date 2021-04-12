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
	"mir-go/daemon/common"
	"mir-go/daemon/fw"
	"mir-go/daemon/lf"
	"mir-go/daemon/plugin"
	"mir-go/daemon/utils"
	"net"
	"net/http"
	_ "net/http/pprof"
	"testing"
	"time"
)

func TestEthernetTransport_Send(t *testing.T) {
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
	time.Sleep(5 * time.Second)
	str := "00:0c:29:fa:de:18"
	remote := "00:0c:29:a1:35:bf"

	var forWarder fw.Forwarder
	var pluginManager plugin.GlobalPluginManager
	forWarder.Init(&pluginManager, blockQueue)

	forWarder.Start()

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
	remoteMac, err1 := net.ParseMAC(remote)
	if err1 != nil {
		fmt.Println("local mac", err1)
	}
	faceid, faceErr := lf.CreateEtherLogicFace("ens33", remoteMac)
	if faceErr != nil {
		common.LogError(faceErr)
	}
	logicFace := LfTb.GetLogicFacePtrById(faceid)
	//指定pprof对外提供的http服务的ip和端口，配置为0.0.0.0表示可以非本机访问
	go func() {
		http.ListenAndServe("0.0.0.0:9999", nil)
	}()
	//fmt.Println(faceid)
	for {
		logicFace.SendInterest(interest)
	}

}
