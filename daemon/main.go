package main

import (
	"mir-go/daemon/common"
	"mir-go/daemon/fw"
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

}
