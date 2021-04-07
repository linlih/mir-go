/**
 * @Author: wzx
 * @Description:
 * @Version: 1.0.0
 * @Date: 2021/3/30 下午6:54
 *@Copyright: MIN-Group；国家重大科技基础设施——未来网络北大实验室；深圳市信息论与未来网络重点实验室
 */

package cli

import (
	"fmt"
	"github.com/urfave/cli"
)

var remote string
var local string
var shcema string
var mtu int

var faceCommands = cli.Command{
	Name:"lf",
	Usage: "logic Face Management",
	Subcommands: []*cli.Command{&faceGetCommands},
}

var faceGetCommands = cli.Command{
	Name: "list",
	Usage:"Show all face info",
	Action: GetAllFace,
	Flags:[]cli.Flag{
		&cli.StringFlag{
			Name: "remote",
			Value: "/localhost",
			Destination: &remote,
		},
		&cli.StringFlag{
			Name: "local",
			Value: "/mgmt",
			Destination: &local,
		},
		&cli.StringFlag{
			Name: "schema",
			Value: "tcp",
			Destination: &shcema,
		},
	},
}

func GetAllFace(c *cli.Context) error{
	fmt.Println("test")
	fmt.Println(remote)
	return nil
}