//
// @Author: weiguohua
// @Description:
// @Version: 1.0.0
// @Date: 2021/3/22 上午11:34
// @Copyright: MIN-Group；国家重大科技基础设施——未来网络北大实验室；深圳市信息论与未来网络重点实验室
//
package lf

import (
	"mir-go/daemon/common"
	"net"
	"os"
	"os/exec"
)

type UnixStreamListener struct {
	listener *net.UnixListener
	filepath string
}

func (u *UnixStreamListener) Init() {
	u.filepath = "/tmp/mirsock"
}

//
// @Description: 创建一个unix类型的logicFace
// @receiver t
// @param conn	新unix scoket连接句柄
//
func (u *UnixStreamListener) createTcpLogicFace(conn net.Conn) {
	logicFace, _ := createUnixLogicFace(conn)
	logicFace.Start()
}

//
// @Description: 接收unix连接，并创建TCP类型的LogicFace
// @receiver t
//
func (u *UnixStreamListener) accept() {
	for true {
		newConnect, err := u.listener.Accept()
		if err != nil {
			common.LogFatal(err)
		}
		u.createTcpLogicFace(newConnect)
	}
}

//
// @Description:  启动监听协程
// @receiver t
//
func (u *UnixStreamListener) Start() {
	err := os.Remove(u.filepath)
	if err != nil {
		common.LogWarn(err)
	}
	addr, err := net.ResolveUnixAddr("unix", u.filepath)
	if err != nil {
		common.LogFatal(err)
		return
	}
	listener, err := net.ListenUnix("unix", addr)
	if err != nil {
		common.LogFatal(err)
		return
	}
	// 设置连接文件的权限为 777 ， 这样主机上其他用户启动的程序才能正常连接
	cmd := exec.Command("/bin/bash", "-c", "chmod 777 "+u.filepath)
	err = cmd.Start()
	if err != nil {
		common.LogFatal(err)
	}
	u.listener = listener
	go u.accept()
}
