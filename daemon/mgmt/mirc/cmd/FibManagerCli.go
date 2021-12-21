// Package cmd
// @Author: Jianming Que
// @Description:
// @Version: 1.0.0
// @Date: 2021/4/19 11:21 上午
// @Copyright: MIN-Group；国家重大科技基础设施——未来网络北大实验室；深圳市信息论与未来网络重点实验室
//
package cmd

import (
	"encoding/json"
	"fmt"
	"github.com/desertbit/grumble"
	"github.com/olekukonko/tablewriter"
	"minlib/common"
	"minlib/component"
	mgmtlib "minlib/mgmt"
	"mir-go/daemon/mgmt"
	"os"
	"strconv"
)

// CreateFibCommands 创建一个 FibCommands
//
// @Description:
// @return grumble.Command
//
func CreateFibCommands(controller *mgmtlib.MIRController) *grumble.Command {
	fc := new(grumble.Command)
	fc.Name = "fib"
	fc.Help = "Fib Management"

	// add
	fc.AddCommand(&grumble.Command{
		Name: "add",
		Help: "Add next hop for specific logic face",
		Args: func(a *grumble.Args) {
			a.String("prefix", "Target identifier prefix")
			a.Uint64("id", "Next hop logic face id")
		},
		Flags: func(f *grumble.Flags) {
			f.Uint64("c", "cost", 0, "Link cost")
		},
		Run: func(c *grumble.Context) error {
			return AddFib(c, controller)
		},
	})

	// del
	fc.AddCommand(&grumble.Command{
		Name: "del",
		Help: "Delete next hop for specific logic face",
		Args: func(a *grumble.Args) {
			a.String("prefix", "Target identifier prefix")
			a.Uint64("id", "Next hop logic face id")
		},
		Run: func(c *grumble.Context) error {
			return DeleteNextHop(c, controller)
		},
	})

	// list
	fc.AddCommand(&grumble.Command{
		Name: "list",
		Help: "Show all fib info",
		Run: func(c *grumble.Context) error {
			return ListFib(c, controller)
		},
	})

	return fc
}

// AddFib 添加下一跳路由
//
// @Description:
// @param c
// @return error
//
func AddFib(c *grumble.Context, controller *mgmtlib.MIRController) error {
	// 解析命令行参数
	prefix := c.Args.String("prefix")
	logicFaceId := c.Args.Uint64("id")
	cost := c.Flags.Uint64("cost")

	parameters := &component.ControlParameters{}
	identifier, err := component.CreateIdentifierByString(prefix)
	if err != nil {
		return err
	}
	parameters.SetPrefix(identifier)
	parameters.SetLogicFaceId(logicFaceId)
	parameters.SetCost(cost)

	// 构造一个命令执行器
	commandExecutor, err := controller.PrepareCommandExecutor(mgmtlib.CreateFibAddCommand(topPrefix, parameters))
	if err != nil {
		return err
	}
	commandExecutor.SetAutoShutdown(true)

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
func DeleteNextHop(c *grumble.Context, controller *mgmtlib.MIRController) error {
	// 解析命令行参数
	prefix := c.Args.String("prefix")
	logicFaceId := c.Args.Uint64("id")

	parameters := &component.ControlParameters{}
	identifier, err := component.CreateIdentifierByString(prefix)
	if err != nil {
		return err
	}
	parameters.SetPrefix(identifier)
	parameters.SetLogicFaceId(logicFaceId)

	// 构造一个命令执行器
	commandExecutor, err := controller.PrepareCommandExecutor(mgmtlib.CreateFibDeleteCommand(topPrefix, parameters))
	if err != nil {
		return err
	}
	commandExecutor.SetAutoShutdown(true)

	// 执行命令
	response, err := commandExecutor.Start()
	if err != nil {
		return err
	}

	// 如果请求成功，则输出结果
	if response.Code == mgmtlib.ControlResponseCodeSuccess {
		common.LogInfo(fmt.Sprintf("Delete next hop for %s => %d success!", prefix, logicFaceId))
	} else {
		// 请求失败，则输出错误信息
		common.LogError(fmt.Sprintf("Delete next hop for %s => %d failed! errMsg: %s", prefix, logicFaceId, response.Msg))
	}
	return nil
}

// ListFib 显示所有前缀对应的所有下一跳信息
//
// @Description:
// @param c
// @return error
//
func ListFib(c *grumble.Context, controller *mgmtlib.MIRController) error {
	// 构造一个命令执行器
	commandExecutor, err := controller.PrepareCommandExecutor(mgmtlib.CreateFibListCommand(topPrefix))
	if err != nil {
		return err
	}
	commandExecutor.SetAutoShutdown(true)

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
