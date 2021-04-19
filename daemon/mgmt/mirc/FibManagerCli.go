// Package main
// @Author: Jianming Que
// @Description:
// @Version: 1.0.0
// @Date: 2021/4/19 11:21 上午
// @Copyright: MIN-Group；国家重大科技基础设施——未来网络北大实验室；深圳市信息论与未来网络重点实验室
//
package main

import (
	"fmt"
	"github.com/urfave/cli/v2"
	"minlib/common"
	"minlib/component"
	mgmtlib "minlib/mgmt"
)

// Fib 控制命令
// @Description:
//
var fibCommands = cli.Command{
	Name:        "fib",
	Usage:       "Fib Management",
	Subcommands: []*cli.Command{&AddFibCommand},
}

// AddFibCommand 添加下一跳命令
// @Description:
//
var AddFibCommand = cli.Command{
	Name:   "add",
	Usage:  "Add next hop for specific logic face, eg.: mirc fib add ",
	Action: AddFib,
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:     "prefix",
			Usage:    "Target identifier",
			Required: true,
		},
		&cli.Uint64Flag{
			Name:     "id",
			Usage:    "Next hop logic face id",
			Required: true,
		},
		&cli.Uint64Flag{
			Name:     "cost",
			Usage:    "Link cost",
			Required: false,
			Value:    0,
		},
	},
}

// AddFib 添加下一跳路由
//
// @Description:
// @param c
// @return error
//
func AddFib(c *cli.Context) error {
	// 解析命令行参数
	prefix := c.String("prefix")
	logicFaceId := c.Uint64("id")
	cost := c.Uint64("cost")

	parameters := &component.ControlParameters{}
	identifier, err := component.CreateIdentifierByString(prefix)
	if err != nil {
		return err
	}
	parameters.SetPrefix(identifier)
	parameters.SetLogicFaceId(logicFaceId)
	parameters.SetCost(cost)

	// 构造一个命令执行器
	controller := GetController()
	commandExecutor, err := controller.PrepareCommandExecutor(mgmtlib.CreateFibAddCommand(topPrefix, parameters))
	if err != nil {
		return err
	}

	// 执行命令
	common.LogInfo("begin start")
	response, err := commandExecutor.Start()
	common.LogInfo("after start")
	if err != nil {
		return err
	}

	// 如果请求成功，则输出结果
	if response.Code == mgmtlib.ControlResponseCodeSuccess {
		common.LogInfo(fmt.Sprintf("Add next hop for %s => %d success!", prefix, logicFaceId))
	} else {
		// 请求失败，则输出错误信息
		common.LogInfo(fmt.Sprintf("Add next hop for %s => %d failed! errMsg: %s", prefix, logicFaceId, response.Msg))
	}
	return nil
}
