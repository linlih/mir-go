package main

import (
	"github.com/urfave/cli"
	"minlib/component"
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
		common.LogInfo(configFilePath)
		mirConfig, err := common.ParseConfig(configFilePath)
		if err != nil {
			common.LogFatal(err)
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
	forwarder.Init(pluginManager, packetQueue)

	// PacketValidator
	packetValidator := new(fw.PacketValidator)
	packetValidator.Init(mirConfig.ParallelVerifyNum, mirConfig.VerifyPacket, packetQueue)

	// LogicFaceSystem
	logicFaceTable := new(lf.LogicFaceTable)
	logicFaceTable.Init()
	logicFaceSystem := new(lf.LogicFaceSystem)
	logicFaceSystem.Init(packetValidator, mirConfig)
	logicFaceSystem.Start()

	// TODO: 在这边启动管理模块的程序
	mgmtSystem := mgmt.CreateMgmtSystem()
	mgmtSystem.SetFIB(forwarder.GetFIB())
	dispatcher := mgmt.CreateDispatcher()
	faceServer, faceClient := lf.CreateInnerLogicFacePair()
	dispatcher.FaceClient = faceClient
	topPrefix, _ := component.CreateIdentifierByString("/min-mir/mgmt/localhost")
	dispatcher.AddTopPrefix(topPrefix)
	mgmtSystem.AddInnerFace(topPrefix, faceServer, 0)
	mgmtSystem.Init(dispatcher, logicFaceSystem.LogicFaceTable())
	dispatcher.Start()

	// 启动转发处理流程（死循环阻塞）
	forwarder.Start()
}
