// Package cmd
// @Author: Jianming Que
// @Description:
// @Version: 1.0.0
// @Date: 2021/4/19 8:49 上午
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
	"sort"
	"strconv"
	"strings"
)

// CreateLogicFaceCommands 创建一个 LogicFaceCommands 命令
//
// @Description:
// @param controller
// @return *LogicFaceCommands
//
func CreateLogicFaceCommands(controller *mgmtlib.MIRController) *grumble.Command {
	lfc := new(grumble.Command)
	lfc.Name = "lf"
	lfc.Help = "Logic Face Management"

	// List
	lfc.AddCommand(&grumble.Command{
		Name: "list",
		Help: "Show all LogicFace",
		Run: func(c *grumble.Context) error {
			return ListLogicFace(c, controller)
		},
	})

	// add
	lfc.AddCommand(&grumble.Command{
		Name: "add",
		Help: "Create new LogicFace",
		Args: func(a *grumble.Args) {
			a.String("remote", "Remote Uri to connect")
			a.String("local", "Local Uri", grumble.Default(""))
		},
		Flags: func(f *grumble.Flags) {
			f.String("p", "persistence", "persist", "Persistence of LogicFace, persist/on-demand")
		},
		Run: func(c *grumble.Context) error {
			return AddLogicFace(c, controller)
		},
	})

	// del
	lfc.AddCommand(&grumble.Command{
		Name: "del",
		Help: "Delete LogicFace",
		Args: func(a *grumble.Args) {
			a.Uint64("id", "The LogicFaceId you need to delete")
		},
		Run: func(c *grumble.Context) error {
			return DelLogicFace(c, controller)
		},
	})

	return lfc
}

// ListLogicFace 获取所有Face信息并展示
//
// @Description:
// @param c
// @return error
//
func ListLogicFace(c *grumble.Context, controller *mgmtlib.MIRController) error {
	commandExecutor, err := controller.PrepareCommandExecutor(mgmtlib.CreateLogicFaceListCommand(topPrefix))
	commandExecutor.SetAutoShutdown(true)
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
func AddLogicFace(c *grumble.Context, controller *mgmtlib.MIRController) error {
	// 从命令行解析参数
	remoteUri := c.Args.String("remote")
	localUri := c.Args.String("local")
	persistency := c.Flags.String("persistence")

	remoteUriItems := strings.Split(remoteUri, "://")
	if len(remoteUriItems) != 2 {
		return FaceManagerCliError{msg: fmt.Sprintf("Remote uri is wrong, expect one '://' item, %s", remoteUri)}
	}
	parameters := new(component.ControlParameters)
	parameters.SetUri(remoteUri)
	parameters.SetUriScheme(uint64(component.GetUriSchemeByString(remoteUriItems[0])))
	if localUri != "" {
		parameters.SetLocalUri(localUri)
	}
	parameters.SetPersistency(uint64(component.GetPersistencyByString(persistency)))

	// 发起一个请求命令得到结果
	commandExecutor, err := controller.PrepareCommandExecutor(mgmtlib.CreateLogicFaceAddCommand(topPrefix, parameters))
	commandExecutor.SetAutoShutdown(true)

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
		common.LogError("Add LogicFace failed, errMsg: ", response.Msg)
	}
	return nil
}

// DelLogicFace 根据 LogicFaceId 删除一个 LogicFace
//
// @Description:
// @param c
// @return error
//
func DelLogicFace(c *grumble.Context, controller *mgmtlib.MIRController) error {
	logicFaceId := c.Args.Uint64("id")

	// 发起一个请求命令得到结果
	commandExecutor, err := controller.PrepareCommandExecutor(mgmtlib.CreateLogicFaceDelCommand(topPrefix, logicFaceId))
	commandExecutor.SetAutoShutdown(true)

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
		common.LogError("Delete LogicFace failed, errMsg: ", response.Msg)
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
