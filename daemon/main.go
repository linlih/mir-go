package main

import (
	"minlib/component"
	"mir-go/daemon/common"
	"mir-go/daemon/fw"
	"mir-go/daemon/lf"
	"mir-go/daemon/mgmt"
	"mir-go/daemon/plugin"
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
	packetQueue := fw.CreateBlockQueue(uint(mirConfig.ForwarderConfig.PacketQueueSize))

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
	logicFaceSystem.Init(logicFaceTable, packetValidator)
	logicFaceSystem.Start()

	// TODO: 在这边启动管理模块的程序
	fibManager := mgmt.CreateFibManager()
	faceManager := mgmt.CreateFaceManager()
	csManager := mgmt.CreateCsManager()
	dispatcher := mgmt.CreateDispatcher()
	fibManager.Init(dispatcher)
	faceManager.Init(dispatcher)
	csManager.Init(dispatcher)
	// FIX:下面这两行暂时保留 后面可能需要删除
	lf.GLogicFaceTable = &lf.LogicFaceTable{}
	lf.GLogicFaceTable.Init()
	faceServer, faceClient := lf.CreateInnerLogicFacePair()
	dispatcher.FaceClient = faceClient
	identifier, err := component.CreateIdentifierByString("/min-mir/mgmt/localhop")
	if err != nil {
		common.LogError("register prefix fail!the err is:", err)
	}
	dispatcher.AddTopPrefix(identifier)
	fibManager.GetFib().AddOrUpdate(identifier, faceServer, 0)
	dispatcher.Start()

	// 启动转发处理流程（死循环阻塞）
	forwarder.Start()
}
