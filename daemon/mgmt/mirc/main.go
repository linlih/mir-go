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
	"mir-go/daemon/mgmt/mirc/cmd"
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

func main() {
	// 首先要求用户输入密码
	passwd := AskPassword()

	controller := cmd.GetController(passwd)

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
