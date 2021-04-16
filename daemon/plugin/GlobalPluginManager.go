// Package plugin
// @Author: Jianming Que
// @Description:
// @Version: 1.0.0
// @Date: 2021/2/21 4:19 下午
// @Copyright: MIN-Group；国家重大科技基础设施——未来网络北大实验室；深圳市信息论与未来网络重点实验室
//
package plugin

import (
	"minlib/component"
	"minlib/packet"
	"mir-go/daemon/lf"
	"mir-go/daemon/table"
)

// GlobalPluginManager
// 全局的插件管理程序，主要包含以下功能
//
// @Description:
//	1. 注册插件；
//	2. 在 Forwarder 处理流程的各个锚点调用对应的回调，在调用时按顺序调用插件列表中对应的回调函数
//
type GlobalPluginManager struct {
	plugins []IPlugin
}

type Task func(plugin IPlugin) int

func (gpm *GlobalPluginManager) doInEveryPlugins(task Task) int {
	result := 0
	for _, plugin := range gpm.plugins {
		result = task(plugin)

		// 如果插件锚点返回值不为0，则拦截后续插件的执行
		if result != 0 {
			break
		}
	}
	return result
}

// RegisterPlugin
// 注册一个插件
//
// @Description:
// @receiver gpm
// @param plugin
//
func (gpm *GlobalPluginManager) RegisterPlugin(plugin IPlugin) {
	gpm.plugins = append(gpm.plugins, plugin)
}

////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
/// 管道锚点
////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

func (gpm *GlobalPluginManager) OnIncomingInterest(ingress *lf.LogicFace, interest *packet.Interest) int {
	return gpm.doInEveryPlugins(func(plugin IPlugin) int {
		return plugin.OnIncomingInterest(ingress, interest)
	})
}

func (gpm *GlobalPluginManager) OnInterestLoop(ingress *lf.LogicFace, interest *packet.Interest) int {
	return gpm.doInEveryPlugins(func(plugin IPlugin) int {
		return plugin.OnInterestLoop(ingress, interest)
	})
}

func (gpm *GlobalPluginManager) OnContentStoreMiss(ingress *lf.LogicFace, pitEntry *table.PITEntry, interest *packet.Interest) int {
	return gpm.doInEveryPlugins(func(plugin IPlugin) int {
		return plugin.OnContentStoreMiss(ingress, pitEntry, interest)
	})
}

func (gpm *GlobalPluginManager) OnContentStoreHit(ingress *lf.LogicFace, pitEntry *table.PITEntry, interest *packet.Interest, data *table.CSEntry) int {
	return gpm.doInEveryPlugins(func(plugin IPlugin) int {
		return plugin.OnContentStoreHit(ingress, pitEntry, interest, data)
	})
}

func (gpm *GlobalPluginManager) OnOutgoingInterest(egress *lf.LogicFace, pitEntry *table.PITEntry, interest *packet.Interest) int {
	return gpm.doInEveryPlugins(func(plugin IPlugin) int {
		return plugin.OnOutgoingInterest(egress, pitEntry, interest)
	})
}

func (gpm *GlobalPluginManager) OnInterestFinalize(pitEntry *table.PITEntry) int {
	return gpm.doInEveryPlugins(func(plugin IPlugin) int {
		return plugin.OnInterestFinalize(pitEntry)
	})
}

func (gpm *GlobalPluginManager) OnIncomingData(ingress *lf.LogicFace, data *packet.Data) int {
	return gpm.doInEveryPlugins(func(plugin IPlugin) int {
		return plugin.OnIncomingData(ingress, data)
	})
}

func (gpm *GlobalPluginManager) OnDataUnsolicited(ingress *lf.LogicFace, data *packet.Data) int {
	return gpm.doInEveryPlugins(func(plugin IPlugin) int {
		return plugin.OnDataUnsolicited(ingress, data)
	})
}

func (gpm *GlobalPluginManager) OnOutgoingData(egress *lf.LogicFace, data *packet.Data) int {
	return gpm.doInEveryPlugins(func(plugin IPlugin) int {
		return plugin.OnOutgoingData(egress, data)
	})
}

func (gpm *GlobalPluginManager) OnIncomingNack(ingress *lf.LogicFace, nack *packet.Nack) int {
	return gpm.doInEveryPlugins(func(plugin IPlugin) int {
		return plugin.OnIncomingNack(ingress, nack)
	})
}

func (gpm *GlobalPluginManager) OnOutgoingNack(egress *lf.LogicFace, pitEntry *table.PITEntry, header *component.NackHeader) int {
	return gpm.doInEveryPlugins(func(plugin IPlugin) int {
		return plugin.OnOutgoingNack(egress, pitEntry, header)
	})
}

func (gpm *GlobalPluginManager) OnIncomingCPacket(ingress *lf.LogicFace, cPacket *packet.CPacket) int {
	return gpm.doInEveryPlugins(func(plugin IPlugin) int {
		return plugin.OnIncomingCPacket(ingress, cPacket)
	})
}

func (gpm *GlobalPluginManager) OnOutgoingCPacket(egress *lf.LogicFace, cPacket *packet.CPacket) int {
	return gpm.doInEveryPlugins(func(plugin IPlugin) int {
		return plugin.OnOutgoingCPacket(egress, cPacket)
	})
}
