/**
 * @Author: wzx
 * @Description:
 * @Version: 1.0.0
 * @Date: 2021/3/30 下午6:54
 *@Copyright: MIN-Group；国家重大科技基础设施——未来网络北大实验室；深圳市信息论与未来网络重点实验室
 */

package cli

import (
	"encoding/json"
	"github.com/urfave/cli"
	common2 "minlib/common"
	"minlib/component"
	"minlib/logicface"
	"minlib/packet"
	"mir-go/daemon/mgmt"
)

var remote string
var local string
var scheme string
var mtu int

var faceCommands = cli.Command{
	Name:        "lf",
	Usage:       "logic Face Management",
	Subcommands: []*cli.Command{&GetFaceInfoCommand},
}

var GetFaceInfoCommand = cli.Command{
	Name:   "list",
	Usage:  "Show all face info",
	Action: GetAllFaceInfo,
}

var SetFaceInfoCommand = cli.Command{
	Name:   "list",
	Usage:  "Show all face info",
	Action: GetAllFaceInfo,
	Flags: []cli.Flag{
		&cli.StringFlag{
			//远端地址
			Name:        "remote",
			Value:       "192.168.1.1",
			Destination: &remote,
		},
		&cli.StringFlag{
			Name:        "local",
			Value:       "127.0.0.1:13899",
			Destination: &local,
		},
		&cli.StringFlag{
			Name:        "schema",
			Value:       "tcp",
			Destination: &scheme,
		},
	},
}

func GetAllFaceInfo(c *cli.Context) error {
	face := &logicface.LogicFace{}
	// 建立unix连接
	err := face.InitWithUnixSocket("/tmp/mirsock")
	if err != nil {
		common2.LogError("connect MIR fail!the err is:", err)
		return err
	}
	interest := &packet.Interest{}
	identifier, _ := component.CreateIdentifierByString("/min-mir/mgmt/localhost/face-mgmt/list")
	interest.SetName(identifier)
	interest.SetTtl(5)
	interest.InterestLifeTime.SetInterestLifeTime(4000)

	if err = face.SendInterest(interest); err != nil {
		common2.LogError("send interest packet fail!the err is:", err)
		return err
	}
	minPacket, err := face.ReceivePacket()
	if err != nil {
		common2.LogError("receive min packet fail!the err is:", err)
		return err
	}
	data, _ := packet.CreateDataByMINPacket(minPacket)
	var respinseHeader *mgmt.ResponseHeader
	err = json.Unmarshal(data.GetValue(), &respinseHeader)
	if err != nil {
		common2.LogError("parse data fail!the err is:", err)
		return err
	}
	for i := 0; i < respinseHeader.FragNums; i++ {

	}
	return nil
}
