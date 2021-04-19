// Package main
// @Author: Jianming Que
// @Description:
// @Version: 1.0.0
// @Date: 2021/4/19 8:49 上午
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
	"sort"
	"strconv"
	"strings"
)

// LogicFace 命令
// @Description:
//
var faceCommands = cli.Command{
	Name:        "lf",
	Usage:       "logic Face Management",
	Subcommands: []*cli.Command{&ListLogicFaceCommand, &AddLogicFaceCommand, &DelLogicFaceCommand},
}

// ListLogicFaceCommand 输出LogicFace列表
// @Description:
//
var ListLogicFaceCommand = cli.Command{
	Name:   "list",
	Usage:  "Show all face info",
	Action: ListLogicFace,
}

// AddLogicFaceCommand 添加一个 LogicFace
// @Description:
//
var AddLogicFaceCommand = cli.Command{
	Name:   "add",
	Usage:  "Create new face",
	Action: AddLogicFace,
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:     "remote",
			Usage:    "remote address for connect",
			Required: true,
		},
		&cli.StringFlag{
			Name:  "local",
			Usage: "local address for accept",
		},
		&cli.Uint64Flag{
			Name:  "mtu",
			Value: 1500,
			Usage: "MTU",
		},
		&cli.StringFlag{
			Name:  "persistency",
			Usage: "Persistency of LogicFace, persist/on-demand",
			Value: "persist",
		},
	},
}

// DelLogicFaceCommand 删除一个 LogicFace
// @Description:
//
var DelLogicFaceCommand = cli.Command{
	Name:   "del",
	Usage:  "Delete face",
	Action: DelLogicFace,
	Flags: []cli.Flag{
		&cli.StringFlag{
			//远端地址
			Name:  "id",
			Value: "",
		},
	},
}

// ListLogicFace 获取所有Face信息并展示
//
// @Description:
// @param c
// @return error
//
func ListLogicFace(c *cli.Context) error {
	controller := GetController()
	commandExecutor, err := controller.PrepareCommandExecutor(mgmtlib.CreateLogicFaceListCommand(topPrefix))
	if err != nil {
		return err
	}

	// 执行命令拉取结果
	response, err := commandExecutor.Start()
	if err != nil {
		return err
	}

	// 反序列化，输出结果
	var faceInfoList []mgmt.FaceInfo
	err = json.Unmarshal(response.GetBytes(), &faceInfoList)
	if err != nil {
		return err
	}

	// 使用表格美化输出
	table := tablewriter.NewWriter(os.Stdout)
	sort.Slice(faceInfoList, func(i, j int) bool {
		return faceInfoList[i].LogicFaceId < faceInfoList[j].LogicFaceId
	})
	for _, v := range faceInfoList {
		table.Append([]string{strconv.FormatUint(v.LogicFaceId, 10), v.LocalUri, v.RemoteUri, strconv.FormatUint(v.Mtu, 10)})
	}
	table.SetHeader([]string{"LogicFaceId", "LocalUri", "RemoteUri", "Mtu"})
	table.SetHeaderColor(
		tablewriter.Colors{tablewriter.FgHiRedColor, tablewriter.Bold},
		tablewriter.Colors{tablewriter.FgHiRedColor, tablewriter.Bold},
		tablewriter.Colors{tablewriter.FgHiRedColor, tablewriter.Bold},
		tablewriter.Colors{tablewriter.FgHiRedColor, tablewriter.Bold})
	table.SetCaption(true, "LogicFace Table Info")
	table.SetAlignment(tablewriter.ALIGN_CENTER)
	table.Render()
	return nil
}

// AddLogicFace 创建一个新的 LogicFace 连接到另一个路由器
//
// @Description:
// @param c
// @return error
//
func AddLogicFace(c *cli.Context) error {
	// 从命令行解析参数
	remoteUri := c.String("remote")
	localUri := c.String("local")
	mtu := c.Uint64("mtu")
	persistency := c.String("persistency")

	remoteUriItems := strings.Split(remoteUri, "://")
	if len(remoteUriItems) != 2 {
		return FaceManagerCliError{msg: fmt.Sprintf("Remote uri is wrong, expect one '://' item, %s", remoteUri)}
	}
	parameters := &component.ControlParameters{}
	parameters.SetUri(remoteUri)
	parameters.SetUriScheme(uint64(component.GetUriSchemeByString(remoteUriItems[0])))
	parameters.SetMtu(mtu)
	parameters.SetLocalUri(localUri)
	parameters.SetPersistency(uint64(component.GetPersistencyByString(persistency)))

	// 发起一个请求命令得到结果
	controller := GetController()
	commandExecutor, err := controller.PrepareCommandExecutor(mgmtlib.CreateLogicFaceAddCommand(topPrefix, parameters))
	if err != nil {
		return err
	}
	response, err := commandExecutor.Start()
	if err != nil {
		return err
	}

	// 如果请求成功，则输出结果
	if response.Code == mgmtlib.ControlResponseCodeSuccess {
		common.LogInfo("Add LogicFace success, id =", response.GetString())
	} else {
		// 请求失败，则输出错误信息
		common.LogInfo("Add LogicFace failed, errMsg: ", response.Msg)
	}
	return nil
}

// DelLogicFace 根据 LogicFaceId 删除一个 LogicFace
//
// @Description:
// @param c
// @return error
//
func DelLogicFace(c *cli.Context) error {
	logicFaceId := c.Uint64("id")

	// 发起一个请求命令得到结果
	controller := GetController()
	commandExecutor, err := controller.PrepareCommandExecutor(mgmtlib.CreateLogicFaceDelCommand(topPrefix, logicFaceId))
	if err != nil {
		return err
	}
	response, err := commandExecutor.Start()
	if err != nil {
		return err
	}

	// 如果请求成功，则输出结果
	if response.Code == mgmtlib.ControlResponseCodeSuccess {
		common.LogInfo("Delete LogicFace success!")
	} else {
		// 请求失败，则输出错误信息
		common.LogInfo("Delete LogicFace failed, errMsg: ", response.Msg)
	}
	return nil
}

/////////////////////////////////////////////////////////////////////////////////////////////////////////
///// 错误处理
/////////////////////////////////////////////////////////////////////////////////////////////////////////

type FaceManagerCliError struct {
	msg string
}

func (f FaceManagerCliError) Error() string {
	return fmt.Sprintf("FaceManagerCliError: %s", f.msg)
}
