package main

import (
	"mir-go/daemon/fw"
	"mir-go/daemon/plugin"
)

func main() {
	InitForwarder()
}

func InitForwarder() {
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
