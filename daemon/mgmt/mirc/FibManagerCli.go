// Package main
// @Author: Jianming Que
// @Description:
// @Version: 1.0.0
// @Date: 2021/4/19 11:21 上午
// @Copyright: MIN-Group；国家重大科技基础设施——未来网络北大实验室；深圳市信息论与未来网络重点实验室
//
package main

import (
	"encoding/json"
	"fmt"
	"github.com/olekukonko/tablewriter"
	"github.com/urfave/cli/v2"
	"minlib/common"
	"minlib/component"
	mgmtlib "minlib/mgmt"
	"mir-go/daemon/mgmt"
	"os"
	"strconv"
)

// Fib 控制命令
// @Description:
//
var fibCommands = cli.Command{
	Name:        "fib",
	Usage:       "Fib Management",
	Subcommands: []*cli.Command{&AddFibCommand, &DeleteNextHopCommand, &ListFibCommand},
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

// DeleteNextHopCommand 删除指定前缀的下一跳命令
// @Description:
//
var DeleteNextHopCommand = cli.Command{
	Name:   "del",
	Usage:  "delete next hop for specific logic face, eg.: mirc fib del ",
	Action: DeleteNextHop,
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
	},
}

// ListFibCommand 展示所有Fib表项命令
// @Description:
//
var ListFibCommand = cli.Command{
	Name:   "list",
	Usage:  "show all fib info, eg.: mirc fib list ",
	Action: ListFib,
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
	response, err := commandExecutor.Start()
	if err != nil {
		return err
	}

	// 如果请求成功，则输出结果
	if response.Code == mgmtlib.ControlResponseCodeSuccess {
		common.LogInfo(fmt.Sprintf("Add next hop for %s => %d success!", prefix, logicFaceId))
	} else {
		// 请求失败，则输出错误信息
		common.LogError(fmt.Sprintf("Add next hop for %s => %d failed! errMsg: %s", prefix, logicFaceId, response.Msg))
	}
	return nil
}

// DeleteNextHop  刪除一个到指定前缀的路由
//
// @Description:
// @param c
// @return error
//
func DeleteNextHop(c *cli.Context) error {
	// 解析命令行参数
	prefix := c.String("prefix")
	logicFaceId := c.Uint64("id")

	parameters := &component.ControlParameters{}
	identifier, err := component.CreateIdentifierByString(prefix)
	if err != nil {
		return err
	}
	parameters.SetPrefix(identifier)
	parameters.SetLogicFaceId(logicFaceId)

	// 构造一个命令执行器
	controller := GetController()
	commandExecutor, err := controller.PrepareCommandExecutor(mgmtlib.CreateFibDeleteCommand(topPrefix, parameters))
	if err != nil {
		return err
	}

	// 执行命令
	response, err := commandExecutor.Start()
	if err != nil {
		return err
	}

	// 如果请求成功，则输出结果
	if response.Code == mgmtlib.ControlResponseCodeSuccess {
		common.LogInfo(fmt.Sprintf("Add next hop for %s => %d success!", prefix, logicFaceId))
	} else {
		// 请求失败，则输出错误信息
		common.LogError(fmt.Sprintf("Add next hop for %s => %d failed! errMsg: %s", prefix, logicFaceId, response.Msg))
	}
	return nil
}

// ListFib 显示所有前缀对应的所有下一跳信息
//
// @Description:
// @param c
// @return error
//
func ListFib(c *cli.Context) error {
	// 构造一个命令执行器
	controller := GetController()
	commandExecutor, err := controller.PrepareCommandExecutor(mgmtlib.CreateFibListCommand(topPrefix))
	if err != nil {
		return err
	}

	// 执行命令
	response, err := commandExecutor.Start()
	if err != nil {
		return err
	}

	// 反序列化，输出结果
	var fibInfoList []mgmt.FibInfo
	err = json.Unmarshal(response.GetBytes(), &fibInfoList)
	if err != nil {
		return err
	}

	// 使用表格美化输出
	table := tablewriter.NewWriter(os.Stdout)

	for _, fibInfo := range fibInfoList {
		for _, nextHopInfo := range fibInfo.NextHopsInfo {
			table.Append([]string{fibInfo.Identifier, strconv.FormatUint(nextHopInfo.LogicFaceId, 10), strconv.FormatUint(nextHopInfo.Cost, 10)})
		}
	}
	table.SetHeader([]string{"Prefix", "LogicFaceId", "Cost"})
	table.SetHeaderColor(
		tablewriter.Colors{tablewriter.FgHiRedColor, tablewriter.Bold},
		tablewriter.Colors{tablewriter.FgHiRedColor, tablewriter.Bold},
		tablewriter.Colors{tablewriter.FgHiRedColor, tablewriter.Bold})
	table.SetCaption(true, "Fib Table Info")
	table.SetAlignment(tablewriter.ALIGN_CENTER)
	table.SetAutoMergeCellsByColumnIndex([]int{0})
	table.Render()
	return nil
}
