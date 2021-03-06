// Copyright [2022] [MIN-Group -- Peking University Shenzhen Graduate School Multi-Identifier Network Development Group]
//
// Licensed under the Apache License, Version 2.0 (the "License"): you may
// not use this file except in compliance with the License. You may obtain
// a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS, WITHOUT
// WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the
// License for the specific language governing permissions and limitations
// under the License.

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
	"minlib/utils"
	common2 "mir-go/daemon/common"
	"mir-go/daemon/mgmt/mirc/cmd"
	"os"
)

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

		// mirc 日志只输出到终端
		mirConfig.LogFilePath = ""
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

	// 要求用户输入密码
	passwd, err := cmd.AskPassword()
	passwd = utils.GetEncryptPasswd(passwd)
	if err != nil {
		common.LogFatal(err)
	}

	if identity := keyChain.GetIdentityByName(mirConfig.GeneralConfig.DefaultId); identity == nil {
		common.LogFatal("Identity => "+mirConfig.GeneralConfig.DefaultId, "not exists")
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
	// 添加 Identity 管理命令
	app.AddCommand(cmd.CreateIdentityCommands(controller))

	grumble.Main(app)
}
