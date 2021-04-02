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
	//data, err := json.Marshal(mirConfig)
	//println(string(data))
	InitForwarder(mirConfig)
}

func InitForwarder(mirConfig *common.MIRConfig) {
	// 初始化日志模块
	common.InitLogger(mirConfig)

	common.LogInfo("hhhhhh")

	// 初始化插件管理器
	pluginManager := new(plugin.GlobalPluginManager)
	// TODO: 在这边注册插件
	//pluginManager.RegisterPlugin()

	// BlockQueue
	// TODO: 读取配置文件设置包队列大小
	packetQueue := fw.CreateBlockQueue(100)

	// 初始化转发器
	forwarder := new(fw.Forwarder)
	forwarder.Init(pluginManager, packetQueue)

}
