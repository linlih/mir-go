package main

import (
	"github.com/AlecAivazis/survey/v2"
	"github.com/urfave/cli/v2"
	common2 "minlib/common"
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
