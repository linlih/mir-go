// Package lf
// @Author: weiguohua
// @Description:
// @Version: 1.0.0
// @Date: 2021/3/14 下午10:04
// @Copyright: MIN-Group；国家重大科技基础设施——未来网络北大实验室；深圳市信息论与未来网络重点实验室
//
package lf

import (
	common2 "minlib/common"
	"mir-go/daemon/common"
	"time"
)

//
// @Description: 全局logicFace系统
//
var gLogicFaceSystem *LogicFaceSystem

// LogicFaceSystem
// @Description: 启动所有类型的Face监听
// 整个 LogicFaceSystem 的工作原理，
//
type LogicFaceSystem struct {
	ethernetListener      EthernetListener
	tcpListener           TcpListener
	udpListener           UdpListener
	unixListener          UnixStreamListener
	logicFaceTable        *LogicFaceTable
	packetValidator       IPacketValidator
	config                *common.MIRConfig
	cleanLogicFaceTimeVal int
}

func (l *LogicFaceSystem) LogicFaceTable() *LogicFaceTable {
	return l.logicFaceTable
}

// Init
// @Description: 初始化LogicFaceSystem对象
// @receiver l
// @param table
//
func (l *LogicFaceSystem) Init(packetValidator IPacketValidator, config *common.MIRConfig) {
	var logicFaceTable LogicFaceTable
	logicFaceTable.Init()
	l.logicFaceTable = &logicFaceTable
	l.packetValidator = packetValidator
	l.config = config
	l.ethernetListener.Init(config.EtherRoutineNumber)
	l.tcpListener.Init(config)
	l.udpListener.Init(config)
	l.unixListener.Init(config)

	l.cleanLogicFaceTimeVal = config.CleanLogicFaceTableTimeVal

	gLogicFaceSystem = l
	logicFaceMaxIdolTimeMs = int64(config.LogicFaceIdleTime)
}

// Start
// @Description: 启动所有类型的Face监听,启用logicFace的清理协程
//		清理协程的工作机制是：每隔300秒扫描一篇logicFaceTable中的Face，如果logicFace在状态等于false，或者logicFace的超时时间已经过期，
//		则清理logicFace。
// @receiver l
//
func (l *LogicFaceSystem) Start() {
	l.ethernetListener.Start()
	if l.config.SupportTCP {
		l.tcpListener.Start()
	}
	if l.config.SupportUDP {
		l.udpListener.Start()
	}
	if l.config.SupportUnix {
		l.unixListener.Start()
	}
	go l.faceCleaner()
}

func (l *LogicFaceSystem) destroyFace(logicFaceId uint64, logicFace *LogicFace) {
	if logicFace.logicFaceType == LogicFaceTypeUDP {
		l.udpListener.DeleteLogicFace(logicFace.transport.GetRemoteAddr())
	} else if logicFace.logicFaceType == LogicFaceTypeEther {
		l.ethernetListener.DeleteLogicFace(logicFace.transport.GetLocalAddr(), logicFace.transport.GetRemoteAddr())
	}
	l.logicFaceTable.RemoveByLogicFaceId(logicFaceId)
}

//
// @Description: 	遍历faceTable，清除过期或失效的logicFace
// @receiver l
//
func (l *LogicFaceSystem) doFaceClean() {
	curTime := getTimestampMS()
	for k, v := range l.logicFaceTable.mLogicFaceTable {
		if v.state == false {
			common2.LogInfo("1. remove logicface id = ", v.LogicFaceId)
			l.destroyFace(k, v)
		} else if v.expireTime < curTime && v.Persistence == 0 { // logicFace已经超时
			common2.LogInfo("2. remove logicface id = ", v.LogicFaceId)
			v.Shutdown()        // 调用shutdown关闭logicFace
			l.destroyFace(k, v) // 将logicFace从全局logicFaceTable中删除
		}
	}
}

//
// @Description:  由协程调用，每300秒执行一个清表操作
// @receiver l
//
func (l *LogicFaceSystem) faceCleaner() {
	for true {
		l.doFaceClean()
		time.Sleep(time.Second * time.Duration(l.cleanLogicFaceTimeVal))
		common2.LogInfo("clean logic face table ---------------------------- ")
	}
}

//
// @Description: 	获取当前unix时间， 单位是 ms
// @return int64
//
func getTimestampMS() int64 {
	curTime := time.Now().UnixNano() / 1000000
	return curTime
}
