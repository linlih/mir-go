//
// @Author: weiguohua
// @Description:
// @Version: 1.0.0
// @Date: 2021/3/16 上午11:37
// @Copyright: MIN-Group；国家重大科技基础设施——未来网络北大实验室；深圳市信息论与未来网络重点实验室
//
package lf

import (
	common2 "minlib/common"
	"net"
	"strconv"
)

//
// @Description:  TCP端口监听器，用于接收远程mir的TCP连接请求，为新连接创建
//			并启动一个TCP-Transport类型的LogicFace
//
type TcpListener struct {
	TcpPort  uint16       // TCP端口号
	listener net.Listener // TCP监听句柄
}

//
// @Description: 	初始化TCP监听器
// @receiver t
// @param logicFaceTable  全局logicFace表指针
//
func (t *TcpListener) Init() {
	t.TcpPort = 13899
}

//
// @Description: 创建一个TCP类型的logicFace
// @receiver t
// @param conn	新TCP连接句柄
//
func (t *TcpListener) tryCreateTcpLogicFace(conn net.Conn) {
	logicFace, _ := createTcpLogicFace(conn)
	logicFace.Start()
}

//
// @Description: 接收TCP连接，并创建TCP类型的LogicFace
// @receiver t
//
func (t *TcpListener) accept() {
	for true {
		newConnect, err := t.listener.Accept()
		if err != nil {
			common2.LogFatal(err)
		}
		t.tryCreateTcpLogicFace(newConnect)
	}
}

//
// @Description:  启动监听协程
// @receiver t
//
func (t *TcpListener) Start() {
	listener, err := net.Listen("tcp", "0.0.0.0:"+strconv.Itoa(int(t.TcpPort)))
	if err != nil {
		common2.LogFatal(err)
		return
	}
	t.listener = listener
	go t.accept()
}
