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

// Package lf
// @Author: weiguohua
// @Description:
// @Version: 1.0.0
// @Date: 2021/3/23 上午10:46
// @Copyright: MIN-Group；国家重大科技基础设施——未来网络北大实验室；深圳市信息论与未来网络重点实验室
//
package lf

import (
	"errors"
	common2 "minlib/common"
	"minlib/logicface"
	"net"
)

//
// @Description: 外部接口，为其他模块提供创建各种接口的函数
//

// CreateEtherLogicFace
// @Description:	给其他模块调用，创建一个Ether类型的LogicFace，传入本地网卡地址和目的MAC地址。
//					本函数会：
//					（1）判断网卡是否存在，网卡是否已经启动； 如果网卡没启动，则返回错误信息；
//					（2）判断 该本地网卡和目的MAC 对应的logicFace 是否已经存在，如果已经存在，则返回已经存在的logicFaceId；
//					（3） 如果不存在，调用内部接口， 创建一个只用于发送数据的LogicFace。
//					注意 ： 本函数不会判断传输的对端MAC地址是否真实存在。
// @param localIfName	本地网卡名
// @param remoteMacAddr		对端MAC地址，如 34:cf:f6:f8:6a:d8
// @return uint64		logicFaceId
// @return error		错误信息
//
func CreateEtherLogicFace(localIfName string, remoteMacAddr net.HardwareAddr) (*LogicFace, error) {
	ifListener := gLogicFaceSystem.ethernetListener.mInterfaceListeners.LoadInterfaceListener(localIfName)
	if ifListener == nil {
		return nil, errors.New("can not find local dev name : " + localIfName)
	}
	logicFace := ifListener.GetLogicFaceByMacAddr(remoteMacAddr.String())
	if logicFace != nil {
		return logicFace, nil
	}
	logicFace, _ = createEtherLogicFace(localIfName, ifListener.macAddr, remoteMacAddr, ifListener.mtu)
	if logicFace == nil {
		return nil, errors.New("create ether logic face fail")
	}
	ifListener.AddLogicFace(remoteMacAddr.String(), logicFace)
	return logicFace, nil
}

// CreateTcpLogicFace
// @Description:  给其他模块调用，创建一个TCP类型的LogicFace，传入对方的TCP地址，格式是 "<ip>:<port>"，如"192.168.3.7:13899"。
//				函数会执行以下操作：
//				（1） 尝试连接远程TCP地址，如果连接不成功，则返回连接错误信息
//				（2） 如果连接成功，调用内部函数，创建一个TCP类型的logicFace
//				（3） 启动该logicFace的接收数据协程
// @param remoteUri		对方的TCP地址，格式是 "<ip>:<port>"，如"192.168.3.7:13899"
// @return uint64		logicFaceId
// @return error		错误信息
//
func CreateTcpLogicFace(remoteUri string, persistency uint64) (*LogicFace, error) {
	conn, err := net.Dial("tcp", remoteUri)
	if err != nil {
		common2.LogWarn(err)
		return nil, err
	}
	logicFace, _ := createTcpLogicFace(conn, persistency)
	return logicFace, nil
}

// CreateUdpLogicFace
// @Description:	给其他模块调用，创建一个UDP类型的LogicFace，传入对方的UDP地址，格式是 "<ip>:<port>"，如"192.168.3.7:13899"
//				函数会执行以下操作：
//				（1） 尝试解析UDP地址，如果解析不成功，则返回连接错误信息
//				（2） 如果解析UDP地址成功，调用内部函数，创建一个UDP类型的logicFace
//				（3） 启动该logicFace的接收数据协程
// @param remoteUri
// @return uint64
// @return error
//
func CreateUdpLogicFace(remoteUri string) (*LogicFace, error) {
	logicFace := gLogicFaceSystem.udpListener.GetLogicFaceByRemoteUri(remoteUri)
	if logicFace != nil {
		return logicFace, nil
	}
	udpAddr, err := net.ResolveUDPAddr("udp4", remoteUri)
	if err != nil {
		common2.LogWarn(err)
		return nil, err
	}
	udpConn, err := net.ListenUDP("udp", nil)
	if err != nil {
		return nil, err
	}
	logicFace, _ = createUdpLogicFace(udpConn, udpAddr)
	gLogicFaceSystem.udpListener.AddLogicFace(remoteUri, logicFace)
	return logicFace, nil
}

// CreateUnixLogicFace
// @Description:  给其他模块调用，创建一个unix socket类型的LogicFace，传入对方的unix地址，格式是 文件路径，如"/tmp/mirsock"。
//				函数会执行以下操作：
//				（1） 尝试解析unix地址，如果解析不成功，则返回连接错误信息
//				（2） 尝试连接远程地址，如果连接不成功，则返回连接错误信息
//				（3） 如果连接成功，调用内部函数，创建一个unix类型的logicFace
//				（4） 启动该logicFace的接收数据协程
// @param remoteUri		传入对方的unix地址，格式是 文件路径，如"/tmp/mirsock"。
// @return uint64		logicFaceId
// @return error		错误信息
//
func CreateUnixLogicFace(remoteUri string) (*LogicFace, error) {
	addr, err := net.ResolveUnixAddr("unix", remoteUri)
	if err != nil {
		common2.LogWarn(err)
		return nil, err
	}
	conn, err := net.DialUnix("unix", nil, addr)
	if err != nil {
		panic("DialUnix failed.")
	}
	logicFace, _ := createUnixLogicFace(conn)
	return logicFace, nil
}

// CreateInnerLogicFacePair
// @Description: 创建一对相互收发包的内部logicFace，　需要调用者自己把要收包的logicface start 起来
// @return *LogicFace	 转发器使用的logicFace
// @return *logicface.LogicFace	其它模使用的logicFace
// @return *
//
func CreateInnerLogicFacePair() (*LogicFace, *logicface.LogicFace) {
	lfServer, lfClient := createInnerLogicFacePair()
	return lfServer, lfClient
}
