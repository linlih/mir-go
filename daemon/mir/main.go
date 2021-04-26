package main

import (
	"github.com/AlecAivazis/survey/v2"
	"github.com/urfave/cli/v2"
	common2 "minlib/common"
	"minlib/component"
	"minlib/security"
	"mir-go/daemon/common"
	"mir-go/daemon/fw"
	"mir-go/daemon/lf"
	"mir-go/daemon/mgmt"
	"mir-go/daemon/plugin"
	"mir-go/daemon/utils"
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

		// 初始化日志模块
		common.InitLogger(mirConfig)
		InitForwarder(mirConfig)
		return nil
	}

	if err := mirApp.Run(os.Args); err != nil {
		return
	}
}

func InitForwarder(mirConfig *common.MIRConfig) {
	// 初始化插件管理器
	pluginManager := new(plugin.GlobalPluginManager)
	// TODO: 在这边注册插件
	//pluginManager.RegisterPlugin()

	// 初始化 BlockQueue
	packetQueue := utils.CreateBlockQueue(uint(mirConfig.ForwarderConfig.PacketQueueSize))

	// 初始化转发器
	forwarder := new(fw.Forwarder)
	if err := forwarder.Init(pluginManager, packetQueue); err != nil {
		common2.LogFatal(err)
	}

	// PacketValidator
	packetValidator := new(fw.PacketValidator)
	packetValidator.Init(mirConfig.ParallelVerifyNum, mirConfig.VerifyPacket, packetQueue)

	// LogicFaceSystem
	logicFaceSystem := new(lf.LogicFaceSystem)
	logicFaceSystem.Init(packetValidator, mirConfig)

	// 管理模块
	faceServer, faceClient := lf.CreateInnerLogicFacePair()
	mgmtSystem := mgmt.CreateMgmtSystem()
	mgmtSystem.SetFIB(forwarder.GetFIB())
	mgmtSystem.BindFibCleaner(logicFaceSystem.LogicFaceTable())
	dispatcher := mgmt.CreateDispatcher(mirConfig)
	InitKeyChain(&dispatcher.KeyChain, mirConfig)
	dispatcher.FaceClient = faceClient
	topPrefix, _ := component.CreateIdentifierByString("/min-mir/mgmt/localhost")
	dispatcher.AddTopPrefix(topPrefix, forwarder.GetFIB(), faceServer)
	mgmtSystem.Init(dispatcher, logicFaceSystem.LogicFaceTable())

	// 启动 LogicFaceSystem
	logicFaceSystem.Start()
	// 启动命令分发程序
	dispatcher.Start()
	// 启动转发处理流程（死循环阻塞）
	forwarder.Start()
}

func askInputPassword() string {
	passwd := ""
	prompt := &survey.Password{
		Message: "Please type your password",
	}
	_ = survey.AskOne(prompt, &passwd)
	return passwd
}

func askSetPasswd(name string) string {
	for true {
		passwd := ""
		prompt := &survey.Password{
			Message: "Please set passwd for " + name,
		}
		_ = survey.AskOne(prompt, &passwd)
		rePasswd := ""
		prompt = &survey.Password{
			Message: "Please confirm your passwd",
		}
		_ = survey.AskOne(prompt, &rePasswd)

		if passwd == rePasswd {
			return passwd
		} else {
			common2.LogError("The two passwords are inconsistent！")
		}
	}
	return ""
}

// InitKeyChain 初始化秘钥链
//
// @Description:
// @param keyChain
//
func InitKeyChain(keyChain *security.KeyChain, config *common.MIRConfig) {
	common2.LogInfo("DB:", utils.GetRelPath(config.SecurityConfig.IdentityDBPath))
	// 初始化KeyChain
	if err := keyChain.InitialKeyChainByPath(utils.GetRelPath(config.SecurityConfig.IdentityDBPath)); err != nil {
		common2.LogFatal(err)
	}

	// 判断指定的网络身份是否存在
	if keyChain.ExistIdentity(config.GeneralConfig.DefaultId) {
		// 存在则要求用户输入密码解锁网络身份
		passwd := askInputPassword()
		if identity := keyChain.GetIdentifyByName(config.GeneralConfig.DefaultId); identity != nil {
			common2.LogDebug(1, identity.IsLocked(), identity.Prikey, identity.PrikeyRawByte)
			if err := keyChain.SetCurrentIdentity(identity, passwd); err != nil {
				common2.LogFatal(err)
			}
			common2.LogDebug(identity.IsLocked(), identity.Prikey, identity.PrikeyRawByte)
		} else {
			common2.LogFatal("identify must not be nil!")
		}
	} else {
		// 不存在则创建新的网络身份，并让用户为该网络身份设置密码
		passwd := askSetPasswd(config.GeneralConfig.DefaultId)
		if identity, err := keyChain.CreateIdentityByName(config.GeneralConfig.DefaultId, passwd); err != nil {
			common2.LogFatal(err)
		} else if err := keyChain.SetCurrentIdentity(identity, passwd); err != nil {
			common2.LogFatal(err)
		}
	}
}
