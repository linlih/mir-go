package main

import (
	"minlib/component"
	"mir-go/daemon/common"
	"mir-go/daemon/fw"
	"mir-go/daemon/lf"
	"mir-go/daemon/mgmt"
	"mir-go/daemon/plugin"
	"mir-go/daemon/utils"
)

func main() {
	mirConfig, err := common.ParseConfig("/usr/local/etc/mir/mirconf.ini")
	if err != nil {
		common.LogFatal(err)
	}

	// 初始化日志模块
	common.InitLogger(mirConfig)
	InitForwarder(mirConfig)
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

	// get LogicFaceTable
	//logicFaceSystem.LogicFaceTable()

	// TODO: 在这边启动管理模块的程序
	fibManager := mgmt.CreateFibManager()
	faceManager := mgmt.CreateFaceManager()
	csManager := mgmt.CreateCsManager()
	dispatcher := mgmt.CreateDispatcher()
	fibManager.Init(dispatcher, logicFaceSystem.LogicFaceTable())
	faceManager.Init(dispatcher, logicFaceSystem.LogicFaceTable())
	csManager.Init(dispatcher, logicFaceSystem.LogicFaceTable())
	faceServer, faceClient := lf.CreateInnerLogicFacePair()
	dispatcher.FaceClient = faceClient
	identifier, _ := component.CreateIdentifierByString("/min-mir/mgmt/localhop")
	dispatcher.AddTopPrefix(identifier)
	fibManager.GetFib().AddOrUpdate(identifier, faceServer, 0)
	dispatcher.Start()

	// 启动转发处理流程（死循环阻塞）
	forwarder.Start()
}
