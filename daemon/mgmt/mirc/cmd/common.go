// Package mir Package main
// @Author: Jianming Que
// @Description:
// @Version: 1.0.0
// @Date: 2021/4/16 7:35 下午
// @Copyright: MIN-Group；国家重大科技基础设施——未来网络北大实验室；深圳市信息论与未来网络重点实验室
//
package cmd

import (
	"github.com/AlecAivazis/survey/v2"
	"minlib/component"
	"minlib/logicface"
	mgmtlib "minlib/mgmt"
	"minlib/packet"
	"minlib/security"
)

// 全局前缀
// @Description:
//
const topPrefix = "/min-mir/mgmt/localhost"

// unix socket 连接地址
// @Description:
//
const unixPath = "/tmp/mir.sock"

// 默认兴趣包生存期
// @Description:
//
const defaultInterestLifetime = 4000

// buildPrefix 构造命令兴趣包请求前缀
//
// @Description:
// @param moduleName
// @param action
// @return string
//
func buildPrefix(moduleName string, action string) string {
	return topPrefix + "/" + moduleName + "/" + action
}

func newCommandInterest(moduleName string, action string) *packet.Interest {
	interest := &packet.Interest{}
	identifier, _ := component.CreateIdentifierByString(buildPrefix(moduleName, action))
	interest.SetName(identifier)
	interest.SetTtl(2)
	interest.InterestLifeTime.SetInterestLifeTime(defaultInterestLifetime)
	interest.IsCommandInterest = true
	return interest
}

// GetController 构造一个通用的用 Unix 通信的本地命令控制器
//
// @Description:
// @return *mgmtlib.MIRController
//
func GetController(keyChain *security.KeyChain) *mgmtlib.MIRController {
	controller := mgmtlib.CreateMIRController(func() (mgmtlib.IMgmtLogicFace, error) {
		face := new(logicface.LogicFace)
		// 建立unix连接
		if err := face.InitWithUnixSocket(unixPath); err != nil {
			return nil, err
		}
		return face, nil
	}, true, keyChain)

	return controller
}

// AskPassword 要求用户输入一个密码
//
// @Description:
// @return string
//
func AskPassword() (string, error) {
	return AskPasswordWithCustomMsg("Please type your password")
}

// AskPasswordWithCustomMsg 要求用户输入一个密码，自定义提示信息
//
// @Description:
// @param msg
// @return string
//
func AskPasswordWithCustomMsg(msg string) (string, error) {
	passwd := ""
	prompt := &survey.Password{
		Message: msg,
	}
	err := survey.AskOne(prompt, &passwd)
	return passwd, err
}
