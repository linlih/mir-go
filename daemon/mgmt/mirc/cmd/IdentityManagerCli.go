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
	"io/ioutil"
	"minlib/common"
	"minlib/component"
	"minlib/mgmt"
	"minlib/minsecurity"
	cert2 "minlib/minsecurity/crypto/cert"
	mgmt2 "mir-go/daemon/mgmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
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

	// importCert
	ic.AddCommand(&grumble.Command{
		Name: mgmt.IdentityManagementActionImportCert,
		Help: "Import cert, contain Name and Public key, can use to verify packet",
		Args: func(a *grumble.Args) {
			a.String("file", "Cert file path")
		},
		Run: func(c *grumble.Context) error {
			return ImportCertIdentity(c, controller)
		},
	})

	// selfIssue
	ic.AddCommand(&grumble.Command{
		Name: mgmt.IdentityManagementActionSelfIssue,
		Help: "Issue cert for self",
		Args: func(a *grumble.Args) {
			a.String("name", "Identity name")
		},
		Run: func(c *grumble.Context) error {
			return SelfIssueIdentity(c, controller)
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
		return err
	}
	parameters.SetPrefix(identifier)

	// 构造一个命令执行器
	commandExecutor, err := controller.PrepareCommandExecutor(mgmt.CreateIdentityDumpCertCommand(topPrefix, parameters))
	if err != nil {
		return err
	}

	// 执行命令
	response, err := commandExecutor.Start()
	if err != nil {
		return err
	}
	if response.Code != mgmt.ControlResponseCodeSuccess {
		common.LogError("Dump cert error =>", response.Msg)
		return nil
	}

	// 反序列化，输出结果
	var identityInfos []string
	err = json.Unmarshal(response.GetBytes(), &identityInfos)
	if err != nil {
		return err
	}

	// 输出
	common.LogInfo(identityInfos[0])

	// 保存文件
	if f, err := os.Create(strings.ReplaceAll(name, "/", "-")[1:] + ".cert"); err != nil {
		common.LogError(err)
	} else {
		defer f.Close()
		if _, err := f.Write([]byte(identityInfos[0])); err != nil {
			common.LogError(err)
		}
		absPath, err := filepath.Abs(f.Name())
		if err != nil {
			common.LogError(err)
		}
		common.LogInfo("Cert file save to:", absPath)
	}
	return nil
}

// ImportCertIdentity 导入网络身份
//
// @Description:
// @param c
// @param controller
// @return error
//
func ImportCertIdentity(c *grumble.Context, controller *mgmt.MIRController) error {
	filePath := c.Args.String("file")
	// 判断文件是否存在
	f, err := os.Open(filePath)
	if err != nil {
		return err
	}

	// 读取文件内容
	data, err := ioutil.ReadAll(f)
	if err != nil {
		return err
	}

	// 尝试本地解析证书，如果本地解析证书失败，就不需要和路由器进行通信了
	cert := cert2.Certificate{}
	if err := cert.FromPem(string(data), nil, minsecurity.SM4ECB); err != nil {
		return err
	}

	// 构造一个命令执行器
	commandExecutor, err := controller.PrepareCommandExecutor(mgmt.CreateIdentityImportCertCommand(topPrefix, filePath))
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
		common.LogInfo(fmt.Sprintf("Load cert success => %s", cert.IssueTo))
	} else {
		common.LogError(fmt.Sprintf("IssueSelf failed => %s", response.Msg))
	}
	return nil
}

// SelfIssueIdentity 某个网络身份给自己签发证书
//
// @Description:
// @param c
// @param controller
// @return error
//
func SelfIssueIdentity(c *grumble.Context, controller *mgmt.MIRController) error {
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
	commandExecutor, err := controller.PrepareCommandExecutor(mgmt.CreateIdentitySelfIssueCommand(topPrefix, parameters))
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
		common.LogInfo(fmt.Sprintf("IssueSelf %s success!", name))
	} else {
		common.LogError(fmt.Sprintf("IssueSelf failed => %s", response.Msg))
	}
	return nil
}
