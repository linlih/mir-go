/**
 * @Author: wzx
 * @Description:
 * @Version: 1.0.0
 * @Date: 2021/3/30 下午6:54
 *@Copyright: MIN-Group；国家重大科技基础设施——未来网络北大实验室；深圳市信息论与未来网络重点实验室
 */

package cli

import (
	"github.com/urfave/cli"
	"minlib/component"
	"minlib/logicface"
	"minlib/packet"
	"mir-go/daemon/common"
)

var remote string
var local string
var shcema string
var mtu int

var faceCommands = cli.Command{
	Name:        "lf",
	Usage:       "logic Face Management",
	Subcommands: []*cli.Command{&GetFaceInfoCommand, &AddFaceCommand, &DeleteFaceCommand},
}

var GetFaceInfoCommand = cli.Command{
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
			Destination: &shcema,
		},
	},
}

var AddFaceCommand = cli.Command{
	Name:   "list",
	Usage:  "Show all face info",
	Action: GetAllFaceInfo,
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:        "remote",
			Value:       "/localhost",
			Destination: &remote,
		},
		&cli.StringFlag{
			Name:        "local",
			Value:       "/mgmt",
			Destination: &local,
		},
		&cli.StringFlag{
			Name:        "schema",
			Value:       "tcp",
			Destination: &shcema,
		},
	},
}

var DeleteFaceCommand = cli.Command{
	Name:   "list",
	Usage:  "Show all face info",
	Action: GetAllFaceInfo,
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:        "remote",
			Value:       "/localhost",
			Destination: &remote,
		},
		&cli.StringFlag{
			Name:        "local",
			Value:       "/mgmt",
			Destination: &local,
		},
		&cli.StringFlag{
			Name:        "schema",
			Value:       "tcp",
			Destination: &shcema,
		},
	},
}

func GetAllFaceInfo(c *cli.Context) error {
	face := &logicface.LogicFace{}
	// 建立unix连接
	err := face.InitWithUnixSocket("/tmp/mirsock")
	if err != nil {
		common.LogError("connect MIR fail!the err is:", err)
		return err
	}
	interest := &packet.Interest{}
	identifier, _ := component.CreateIdentifierByString("/min-mir/mgmt/localhop/fib-mgmt/list")
	interest.SetName(identifier)
	if err = face.SendInterest(interest); err != nil {
		common.LogError("send interest packet fail!the err is:", err)
		return err
	}
	face.ReceivePacket()
	return nil
}
