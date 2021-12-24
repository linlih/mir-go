// Package main
// @Author: Jianming Que
// @Description:
//	1. 本命令行工具主要用于初始化MIR系统
// @Version: 1.0.0
// @Date: 2021/12/24 10:38 AM
// @Copyright: MIN-Group；国家重大科技基础设施——未来网络北大实验室；深圳市信息论与未来网络重点实验室
//
package main

import (
	"errors"
	"flag"
	"fmt"
	"github.com/AlecAivazis/survey/v2"
	"io/ioutil"
	common2 "minlib/common"
	"minlib/security"
	"mir-go/daemon/common"
	mir "mir-go/daemon/mircmd"
	"mir-go/daemon/utils"
	"os"
	"strings"
)

const (
	defaultConfigFilePath = "/usr/local/etc/mir/mirconf.ini" // MIR配置文件路径

)

var (
	// 参数
	resetPasswd     = false
	oldPasswdNoHash = false
)

func init() {
	flag.BoolVar(&resetPasswd, "rp", false, "reset passwd")
	flag.BoolVar(&oldPasswdNoHash, "oldPasswdNoHash", false, "Force old passwd do not use SM3 hash")
}

func main() {
	flag.Parse()

	mirConfig, err := common.ParseConfig(defaultConfigFilePath)
	if err != nil {
		common2.LogFatal(err)
	}
	keyChain, err := security.NewKeyChain()
	if err != nil {
		common2.LogFatal(err)
	}

	passwd := ""
	// 判断网络身份是否存在
	if keyChain.ExistIdentity(mirConfig.GeneralConfig.DefaultId) {
		// 网络身份存在
		if resetPasswd {
			// 重置密码
			if oldPasswd, newPasswd, err := askResetPasswd(mirConfig.GeneralConfig.DefaultId, func(passwd string) error {
				return checkPasswd(keyChain, mirConfig.GeneralConfig.DefaultId, passwd)
			}); err != nil {
				common2.LogFatal(err)
			} else {
				if err := keyChain.ResetIdentityPasswd(mirConfig.GeneralConfig.DefaultId, oldPasswd, newPasswd); err != nil {
					common2.LogFatal(err)
				}
				common2.LogInfo("Change passwd success~")
			}

		} else {
			// 输入密码
			passwd = askInputPassword()
			if err := checkPasswd(keyChain, mirConfig.GeneralConfig.DefaultId, passwd); err != nil {
				common2.LogFatal(err)
			}
		}
	} else {
		// 网络身份不存在
		passwd = askSetPasswd(mirConfig.GeneralConfig.DefaultId)
		if identity, err := keyChain.CreateIdentityByName(mirConfig.GeneralConfig.DefaultId, passwd); err != nil {
			common2.LogFatal(err)
		} else if err := keyChain.SetCurrentIdentity(identity, passwd); err != nil {
			common2.LogFatal(err)
		}

	}

	// 将密码生成后放到特定的位置
	savePath := fmt.Sprintf("%s%cpasswd%s", mirConfig.GeneralConfig.EncryptedPasswdSavePath,
		os.PathSeparator, strings.Replace(mirConfig.GeneralConfig.DefaultId, "/", "-", -1))

	if _, ok := utils.IsFile(savePath); !ok {
		//新建文件
		if newFile, err := os.Create(savePath); err != nil {
			common2.LogFatal(err)
		} else {
			defer newFile.Close()
			if _, err := newFile.WriteString(passwd); err != nil {
				common2.LogFatal(err)
			}
		}
	} else {
		if err := ioutil.WriteFile(savePath, []byte(passwd), 0666); err != nil {
			common2.LogFatal(err)
		}
	}
	common2.LogInfo("Save passwd file success~")
}

// checkPasswd 检查密码是否有效
//
// @Description:
// @param keyChain
// @param name
// @param passwd
// @return error
//
func checkPasswd(keyChain *security.KeyChain, name string, passwd string) error {
	if myIdentity := keyChain.GetIdentityByName(name); myIdentity == nil {
		return errors.New("Identity not exists")
	} else {
		return keyChain.SetCurrentIdentity(myIdentity, passwd)
	}
}

// askInputPassword 要求用户输入密码
//
// @Description:
// @return string
//
func askInputPassword() string {
	passwd := ""
	prompt := &survey.Password{
		Message: "Please type your password",
	}
	if err := survey.AskOne(prompt, &passwd); err != nil {
		common2.LogFatal(err)
	} else {
		// 迅速将明文的密码转为 SM3 hash值
		if !oldPasswdNoHash {
			passwd = mir.GetEncryptPasswd(passwd)
		}
	}
	return passwd
}

// askSetPasswd 要求用户设置密码
//
// @Description:
// @param name
// @return string
//
func askSetPasswd(name string) string {
	for true {
		passwd := ""
		prompt := &survey.Password{
			Message: "Please set passwd for " + name,
		}
		if err := survey.AskOne(prompt, &passwd); err != nil {
			common2.LogFatal(err)
		}
		// 迅速将明文的密码转为 SM3 hash值
		passwd = mir.GetEncryptPasswd(passwd)

		rePasswd := ""
		prompt = &survey.Password{
			Message: "Please confirm your passwd",
		}
		if err := survey.AskOne(prompt, &rePasswd); err != nil {
			common2.LogFatal(err)
		}
		// 迅速将明文的密码转为 SM3 hash值
		rePasswd = mir.GetEncryptPasswd(rePasswd)

		if passwd == rePasswd {
			return passwd
		} else {
			common2.LogError("The two passwords are inconsistent！")
		}
	}
	return ""
}

func askResetPasswd(name string, checkPasswd func(passwd string) error) (string, string, error) {
	for true {
		oldPasswd := ""
		prompt := &survey.Password{
			Message: "Please input old passwd for " + name,
		}
		if err := survey.AskOne(prompt, &oldPasswd); err != nil {
			common2.LogFatal(err)
		}

		if !oldPasswdNoHash {
			oldPasswd = mir.GetEncryptPasswd(oldPasswd)
		}

		// 检查密码是否正确
		if err := checkPasswd(oldPasswd); err != nil {
			common2.LogFatal(err)
		}

		newPasswd := ""
		prompt = &survey.Password{
			Message: "Please set new passwd",
		}
		if err := survey.AskOne(prompt, &newPasswd); err != nil {
			common2.LogFatal(err)
		}
		newPasswd = mir.GetEncryptPasswd(newPasswd)

		rePasswd := ""
		prompt = &survey.Password{
			Message: "Please confirm your passwd",
		}
		if err := survey.AskOne(prompt, &rePasswd); err != nil {
			common2.LogFatal(err)
			continue
		}
		rePasswd = mir.GetEncryptPasswd(rePasswd)

		if newPasswd == rePasswd {
			return oldPasswd, newPasswd, nil
		} else {
			common2.LogFatal("The two passwords are inconsistent！")
		}
	}
	return "", "", nil
}
