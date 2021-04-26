// Package main
// @Author: Jianming Que
// @Description:
// @Version: 1.0.0
// @Date: 2021/4/16 7:31 下午
// @Copyright: MIN-Group；国家重大科技基础设施——未来网络北大实验室；深圳市信息论与未来网络重点实验室
//

package main

import (
	"github.com/AlecAivazis/survey/v2"
	"github.com/desertbit/grumble"
	"github.com/urfave/cli/v2"
	"minlib/common"
	"minlib/security"
	common2 "mir-go/daemon/common"
	"mir-go/daemon/mgmt/mirc/cmd"
	"os"
)

// AskPassword 要求用户输入一个密码
//
// @Description:
// @return string
//
func AskPassword() string {
	passwd := ""
	prompt := &survey.Password{
		Message: "Please type your password",
	}
	_ = survey.AskOne(prompt, &passwd)
	return passwd
}

// AskIdentityName 要求用户输入一个使用的网络身份
//
// @Description:
// @return string
//
func AskIdentityName() string {
	identityName := ""
	prompt := &survey.Input{
		Message: "Please type your identity name",
	}
	_ = survey.AskOne(prompt, &identityName)
	return identityName
}

const defaultConfigFilePath = "/usr/local/etc/mir/mirconf.ini"

func main() {
	var configFilePath string
	mirApp := cli.NewApp()
	mirApp.Name = "mir"
	mirApp.Usage = " MIR forwarder daemon program "
	mirApp.Flags = []cli.Flag{
		&cli.StringFlag{
			Name:        "f",
			Value:       defaultConfigFilePath,
			Usage:       "Config file path for MIR",
			Destination: &configFilePath,
			Required:    true,
		},
	}
	mirApp.Action = func(context *cli.Context) error {
		common.LogInfo(configFilePath)
		mirConfig, err := common2.ParseConfig(configFilePath)
		if err != nil {
			common.LogFatal(err)
		}

		// 初始化日志模块
		common2.InitLogger(mirConfig)
		mirc(mirConfig)
		return nil
	}

	if err := mirApp.Run(os.Args); err != nil {
		return
	}
}

func mirc(mirConfig *common2.MIRConfig) {
	// 创建一个KeyChain，并使用气默认身份进行签名
	keyChain := new(security.KeyChain)
	if err := keyChain.InitialKeyChainByPath(mirConfig.IdentityDBPath); err != nil {
		common.LogFatal(err)
	}

	var identityName = ""

	// 输入网络身份
	identityName = AskIdentityName()
	if !keyChain.ExistIdentity(identityName) {
		common.LogFatal("Identity => "+identityName, "not exists, please try again!")
	}

	// 要求用户输入密码
	passwd := AskPassword()

	if identity := keyChain.GetIdentifyByName(identityName); identity == nil {
		common.LogFatal("Identity => "+identityName, "not exists")
	} else {
		if err := keyChain.SetCurrentIdentity(identity, passwd); err != nil {
			common.LogFatal(err)
		}
	}

	controller := cmd.GetController(keyChain)

	// 创建并启动一个交互式命令行工具
	app := grumble.New(&grumble.Config{
		Name:        "mirc",
		Description: "MIR Management Cli Tools",
	})

	// 添加 LogicFace 管理命令
	app.AddCommand(cmd.CreateLogicFaceCommands(controller))
	// 添加 Fib 管理命令
	app.AddCommand(cmd.CreateFibCommands(controller))

	grumble.Main(app)
}
