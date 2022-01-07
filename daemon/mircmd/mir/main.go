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
// @Date: 2021/12/21 4:32 PM
// @Copyright: MIN-Group；国家重大科技基础设施——未来网络北大实验室；深圳市信息论与未来网络重点实验室
//

package main

import (
	"github.com/AlecAivazis/survey/v2"
	"github.com/urfave/cli/v2"
	common2 "minlib/common"
	"minlib/utils"
	"mir-go/daemon/common"
	"mir-go/daemon/mircmd"
	"os"
)

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
		common2.LogInfo(configFilePath)
		mirConfig, err := common.ParseConfig(configFilePath)
		if err != nil {
			common2.LogFatal(err)
		}

		starter := mir.NewMIRStarter(mirConfig)
		passwd := ""
		if starter.IsExistDefaultIdentity() {
			passwd = askInputPassword()
		} else {
			passwd = askSetPasswd(mirConfig.GeneralConfig.DefaultId)
		}
		passwd = utils.GetEncryptPasswd(passwd)
		starter.Start(passwd)
		return nil
	}

	if err := mirApp.Run(os.Args); err != nil {
		return
	}
}

func askInputPassword() string {
	passwd := ""
	prompt := &survey.Password{
		Message: "Please type your password",
	}
	if err := survey.AskOne(prompt, &passwd); err != nil {
		common2.LogFatal(err)
	}

	return passwd
}

func askSetPasswd(name string) string {
	for true {
		passwd := ""
		prompt := &survey.Password{
			Message: "Please set passwd for " + name,
		}
		if err := survey.AskOne(prompt, &passwd); err != nil {
			common2.LogFatal(err)
		}
		rePasswd := ""
		prompt = &survey.Password{
			Message: "Please confirm your passwd",
		}
		if err := survey.AskOne(prompt, &rePasswd); err != nil {
			common2.LogFatal(err)
		}

		if passwd == rePasswd {
			return passwd
		} else {
			common2.LogError("The two passwords are inconsistent！")
		}
	}
	return ""
}
