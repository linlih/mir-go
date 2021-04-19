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
	"github.com/urfave/cli/v2"
	"minlib/common"
	"minlib/component"
	mgmtlib "minlib/mgmt"
	"mir-go/daemon/mgmt"
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
		common.LogError("parse data fail!the err is:", err)
	}
	for _, v := range faceInfoList {
		fmt.Printf("%+v\n", v)
	}
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
		common.LogFatal("Remote uri is wrong, expect one '://' item, ", remoteUri, remoteUriItems)
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
		fmt.Printf("%+v\n", response)
	} else {
		// 请求失败，则
		fmt.Printf("%+v\n", response)
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
		fmt.Printf("%+v\n", response)
	} else {
		// 请求失败，则
		fmt.Printf("%+v\n", response)
	}
	return nil
}
