// Package cmd
// @Author: Jianming Que
// @Description:
// @Version: 1.0.0
// @Date: 2021/4/27 10:48 上午
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
	"minlib/mgmt"
	mgmt2 "mir-go/daemon/mgmt"
	"os"
	"sort"
)

// CreateIdentityCommands 创建一个 IdentityCommands
//
// @Description:
// @param controller
// @return *grumble.Command
//
func CreateIdentityCommands(controller *mgmt.MIRController) *grumble.Command {
	ic := new(grumble.Command)
	ic.Name = "identity"
	ic.Help = "Identity Management"

	// add
	ic.AddCommand(&grumble.Command{
		Name: mgmt.IdentityManagementActionAdd,
		Help: "Create new Identity",
		Args: func(a *grumble.Args) {
			a.String("name", "Identity name")
		},
		Run: func(c *grumble.Context) error {
			return AddIdentity(c, controller)
		},
	})

	// del
	ic.AddCommand(&grumble.Command{
		Name: mgmt.IdentityManagementActionDel,
		Help: "Delete specific Identity",
		Args: func(a *grumble.Args) {
			a.String("name", "Identity name")
		},
		Run: func(c *grumble.Context) error {
			return DelIdentity(c, controller)
		},
	})

	// list
	ic.AddCommand(&grumble.Command{
		Name: mgmt.IdentityManagementActionList,
		Help: "List all identities",
		Run: func(c *grumble.Context) error {
			return ListIdentity(c, controller)
		},
	})

	// dumpCert
	ic.AddCommand(&grumble.Command{
		Name: mgmt.IdentityManagementActionDumpCert,
		Help: "Dump specific identity's cert",
		Args: func(a *grumble.Args) {
			a.String("name", "Identity name")
		},
		Run: func(c *grumble.Context) error {
			return DumpCertIdentity(c, controller)
		},
	})

	return ic
}

// AddIdentity 添加一个新的网络身份
//
// @Description:
// @param c
// @param controller
// @return error
//
func AddIdentity(c *grumble.Context, controller *mgmt.MIRController) error {
	// 解析命令行参数
	name := c.Args.String("name")

	// 要求用户输入一个密码
	passwd := AskPassword()

	parameters := &component.ControlParameters{}
	identifier, err := component.CreateIdentifierByString(name)
	if err != nil {
		return err
	}
	parameters.SetPrefix(identifier)
	parameters.SetPasswd(passwd)

	// 构造一个命令执行器
	commandExecutor, err := controller.PrepareCommandExecutor(mgmt.CreateIdentityAddCommand(topPrefix, parameters))
	if err != nil {
		return err
	}

	// 执行命令
	response, err := commandExecutor.Start()
	if err != nil {
		return err
	}

	// 如果请求成功，则输出结果
	if response.Code == mgmt.ControlResponseCodeSuccess {
		common.LogInfo(fmt.Sprintf("Create new identity %s success!", name))
	} else {
		common.LogError(fmt.Sprintf("Create new identity failed => %s", response.Msg))
	}
	return nil
}

// DelIdentity 删除一个指定的网络身份
//
// @Description:
// @param c
// @param controller
// @return error
//
func DelIdentity(c *grumble.Context, controller *mgmt.MIRController) error {
	// 解析命令行参数
	name := c.Args.String("name")

	// 要求用户输入一个密码
	passwd := AskPassword()

	parameters := &component.ControlParameters{}
	identifier, err := component.CreateIdentifierByString(name)
	if err != nil {
		return err
	}
	parameters.SetPrefix(identifier)
	parameters.SetPasswd(passwd)

	// 构造一个命令执行器
	commandExecutor, err := controller.PrepareCommandExecutor(mgmt.CreateIdentityDelCommand(topPrefix, parameters))
	if err != nil {
		return err
	}

	// 执行命令
	response, err := commandExecutor.Start()
	if err != nil {
		return err
	}

	// 如果请求成功，则输出结果
	if response.Code == mgmt.ControlResponseCodeSuccess {
		common.LogInfo(fmt.Sprintf("Delete identity %s success! => %s", name, response.Msg))
	} else {
		common.LogError(fmt.Sprintf("Delete identity failed => %s", response.Msg))
	}
	return nil
}

// ListIdentity 列出所有的网络身份
//
// @Description:
// @param c
// @param controller
// @return error
//
func ListIdentity(c *grumble.Context, controller *mgmt.MIRController) error {
	// 构造一个命令执行器
	commandExecutor, err := controller.PrepareCommandExecutor(mgmt.CreateIdentityListCommand(topPrefix))
	if err != nil {
		return err
	}

	// 执行命令
	response, err := commandExecutor.Start()
	if err != nil {
		return err
	}

	// 反序列化，输出结果
	var identityInfos []mgmt2.ListIdentityInfo
	err = json.Unmarshal(response.GetBytes(), &identityInfos)
	if err != nil {
		return err
	}

	// 使用表格美化输出
	table := tablewriter.NewWriter(os.Stdout)

	// 排序
	sort.Slice(identityInfos, func(i, j int) bool {
		return identityInfos[i].Name < identityInfos[j].Name
	})

	for _, identityInfo := range identityInfos {
		table.Append([]string{identityInfo.Name})
	}

	table.SetHeader([]string{"Name"})
	table.SetHeaderColor(tablewriter.Colors{tablewriter.FgHiRedColor, tablewriter.Bold})
	table.SetCaption(true, "Identity Table Info")
	table.SetAlignment(tablewriter.ALIGN_CENTER)
	table.Render()
	return nil
}

// DumpCertIdentity 导出指定网络身份的证书
//
// @Description:
// @param c
// @param controller
// @return error
//
func DumpCertIdentity(c *grumble.Context, controller *mgmt.MIRController) error {
	// 解析命令行参数
	name := c.Args.String("name")

	parameters := &component.ControlParameters{}
	identifier, err := component.CreateIdentifierByString(name)
	if err != nil {
		common.LogFatal(err)
	}
	parameters.SetPrefix(identifier)

	// 构造一个命令执行器
	commandExecutor, err := controller.PrepareCommandExecutor(mgmt.CreateIdentityDumpCertCommand(topPrefix, parameters))
	if err != nil {
		common.LogFatal(err)
	}

	// 执行命令
	response, err := commandExecutor.Start()
	if err != nil {
		common.LogFatal(err)
	}
	if response.Code != mgmt.ControlResponseCodeSuccess {
		common.LogError("Dump cert error =>", response.Msg)
		return nil
	}

	// 反序列化，输出结果
	var identityInfos []string
	err = json.Unmarshal(response.GetBytes(), &identityInfos)
	if err != nil {
		common.LogFatal(err)
	}

	// 输出
	common.LogInfo(identityInfos[0])
	return nil
}
