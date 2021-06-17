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
	"mir-go/daemon/table"
	"mir-go/daemon/utils"
	"net"
	"os"
	"time"
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

	// 加载静态路由配置
	go SetUpDefaultRoute(mirConfig.DefaultRouteConfigPath, forwarder.GetFIB())

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
		if identity := keyChain.GetIdentityByName(config.GeneralConfig.DefaultId); identity != nil {
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

//
// @Description: 加载静态路由配置文件
// @param defaultRouteConfigPath	静态路由配置文件的文件路径
// @param fib	FIB表指针
//
func SetUpDefaultRoute(defaultRouteConfigPath string, fib *table.FIB) {
	time.Sleep(time.Second * 2)
	defaultRouteConfig, err := common.ParseDefaultConfig(defaultRouteConfigPath)
	if err != nil {
		common2.LogError("load default route error: ", err, ", ", defaultRouteConfigPath)
		return
	}
	for i := 0; i < len(defaultRouteConfig.Link); i++ {
		remoteUri := defaultRouteConfig.Link[i].RemoteUri
		var logicFace *lf.LogicFace
		if len(remoteUri) <= 0 {
			common2.LogError("remote uri error: ", remoteUri)
			continue
		}
		if remoteUri[:3] == "udp" {
			logicFace, err = lf.CreateUdpLogicFace(remoteUri[6:])
		} else if remoteUri[:3] == "tcp" {
			logicFace, err = lf.CreateTcpLogicFace(remoteUri[6:])
		} else if remoteUri[:3] == "eth" {
			remoteAddr, err := net.ParseMAC(remoteUri[8:])
			if err != nil {
				common2.LogError("parse mac addr error: ", err)
				continue
			}
			logicFace, err = lf.CreateEtherLogicFace(defaultRouteConfig.Link[i].LocalUri, remoteAddr)
		}
		if logicFace == nil || err != nil {
			common2.LogError("create static logic face error: ", err)
			continue
		}
		common2.LogInfo("create default face: ", logicFace.GetLocalUri(), "->", logicFace.GetRemoteUri(), ", face id = ", logicFace.LogicFaceId)
		logicFace.SetPersistence(uint64(defaultRouteConfig.Link[i].Persistence))
		for j := 0; j < len(defaultRouteConfig.Link[i].Routes.Route); j++ {
			identifier, err := component.CreateIdentifierByString(defaultRouteConfig.Link[i].Routes.Route[j].Identifier)
			if err != nil {
				common2.LogError("create identifier from string error: ", err)
				continue
			}
			fib.AddOrUpdate(identifier, logicFace, uint64(defaultRouteConfig.Link[i].Routes.Route[j].Cost))
			common2.LogInfo("add route prefix=", identifier.ToUri(), " -> logic face id = ", logicFace.LogicFaceId)
		}
	}
}
