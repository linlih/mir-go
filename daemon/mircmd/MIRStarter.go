// Copyright [2022] [MIN-Group -- Peking University Shenzhen Graduate School Multi-Identifier Network Development Group]
//
// Licensed under the Apache License, Version 2.0 (the "License"): you may
// not use this file except in compliance with the License. You may obtain
// a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS, WITHOUT
// WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the
// License for the specific language governing permissions and limitations
// under the License.

// Package mir
// @Author: Jianming Que
// @Description:
// @Version: 1.0.0
// @Date: 2021/12/23 9:13 PM
// @Copyright: MIN-Group；国家重大科技基础设施——未来网络北大实验室；深圳市信息论与未来网络重点实验室
//
package mir

import (
	"errors"
	common2 "minlib/common"
	"minlib/component"
	"minlib/security"
	"minlib/utils"
	"mir-go/daemon/common"
	"mir-go/daemon/fw"
	"mir-go/daemon/lf"
	"mir-go/daemon/mgmt"
	"mir-go/daemon/plugin"
	"mir-go/daemon/table"
	utils2 "mir-go/daemon/utils"
	"net"
	"time"
)

const defaultConfigFilePath = "/usr/local/etc/mir/mirconf.ini"

// MIRStarter MIR 启动器
//
// @Description:
//
type MIRStarter struct {
	plugin.GlobalPluginManager                     // 全局插件管理器
	keyChain                   security.KeyChain   // 秘钥链
	mirConfig                  *common.MIRConfig   // MIR 配置文件
	forwarder                  *fw.Forwarder       //转发器
	logicFaceSystem            *lf.LogicFaceSystem // 管理LogicFace
	dispatcher                 *mgmt.Dispatcher    // 管理命令分发器
}

// NewMIRStarter 新建一个 MIR 启动器
//
// @Description:
// @param mirConfig
// @return *MIRStarter
//
func NewMIRStarter(mirConfig *common.MIRConfig) *MIRStarter {
	mirStarter := new(MIRStarter)
	mirStarter.Init(mirConfig)
	return mirStarter
}

func (m *MIRStarter) Init(mirConfig *common.MIRConfig) {
	m.mirConfig = mirConfig
	// 初始化日志模块
	common.InitLogger(mirConfig)

	// 初始化KeyChain
	if err := m.keyChain.InitialKeyChainByPath(utils2.GetRelPath(m.mirConfig.SecurityConfig.IdentityDBPath)); err != nil {
		common2.LogFatal(err)
	}

	// 初始化 BlockQueue
	packetQueue := utils.NewBlockQueue(uint(m.mirConfig.ForwarderConfig.PacketQueueSize))

	// 初始化转发器
	m.forwarder = new(fw.Forwarder)
	if err := m.forwarder.Init(m.mirConfig, &m.GlobalPluginManager, packetQueue); err != nil {
		common2.LogFatal(err)
	}

	// PacketValidator
	packetValidator := new(fw.PacketValidator)
	packetValidator.Init(m.mirConfig.ParallelVerifyNum, m.mirConfig.VerifyPacket, packetQueue)

	// LogicFaceSystem
	m.logicFaceSystem = new(lf.LogicFaceSystem)
	m.logicFaceSystem.Init(packetValidator, m.mirConfig)

	// 管理模块
	faceServer, faceClient := lf.CreateInnerLogicFacePair()
	mgmtSystem := mgmt.CreateMgmtSystem()
	mgmtSystem.SetFIB(m.forwarder.GetFIB())
	mgmtSystem.BindFibCleaner(m.logicFaceSystem.LogicFaceTable())
	m.dispatcher = mgmt.CreateDispatcher(m.mirConfig, &m.keyChain)
	m.dispatcher.FaceClient = faceClient
	topPrefix, _ := component.CreateIdentifierByString("/min-mir/mgmt/localhost")
	m.dispatcher.AddTopPrefix(topPrefix, m.forwarder.GetFIB(), faceServer)
	mgmtSystem.Init(m.dispatcher, m.logicFaceSystem.LogicFaceTable())

	// 加载静态路由配置
	utils2.GoroutineNoPanic(func() {
		SetUpDefaultRoute(m.mirConfig.DefaultRouteConfigPath, m.forwarder.GetFIB())
	})
}

// Start 传入所使用身份的密码，启动MIR
//
// @Description:
// @param pwd
//
func (m *MIRStarter) Start(pwd string) (string, error) {
	// 初始化秘钥
	if err := m.initKeyChain(pwd); err != nil {
		return "", err
	}

	// 启动 LogicFaceSystem
	m.logicFaceSystem.Start()

	// 启动命令分发程序
	m.dispatcher.Start()
	// 启动转发处理流程（死循环阻塞）
	return m.forwarder.Start()
}

// IsExistDefaultIdentity 判断默认身份是否存在
//
// @Description:
// @return bool
//
func (m *MIRStarter) IsExistDefaultIdentity() bool {
	return m.keyChain.ExistIdentity(m.mirConfig.GeneralConfig.DefaultId)
}

// initKeyChain 初始化秘钥链
//
// @Description:
// @param keyChain
//
func (m *MIRStarter) initKeyChain(passwd string) error {
	// 判断指定的网络身份是否存在
	if m.IsExistDefaultIdentity() {
		// 存在则使用输入的密码解密
		if identity := m.keyChain.GetIdentityByName(m.mirConfig.GeneralConfig.DefaultId); identity != nil {
			if err := m.keyChain.SetCurrentIdentity(identity, passwd); err != nil {
				return err
			}
			common2.LogDebug(identity.IsLocked(), identity.Prikey, identity.PrikeyRawByte)
		} else {
			return errors.New("identify must not be nil!")
		}
	} else {
		// 不存在则创建新的网络身份，并使用用户传入的密码作为它的密码
		if identity, err := m.keyChain.CreateIdentityByName(m.mirConfig.GeneralConfig.DefaultId, passwd); err != nil {
			return err
		} else if err := m.keyChain.SetCurrentIdentity(identity, passwd); err != nil {
			return err
		}
	}
	return nil
}

// SetUpDefaultRoute
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
			logicFace, err = lf.CreateTcpLogicFace(remoteUri[6:], 1)
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
