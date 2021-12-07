// Package lf
// @Author: weiguohua
// @Description:
// @Version: 1.0.0
// @Date: 2021/3/16 上午11:33
// @Copyright: MIN-Group；国家重大科技基础设施——未来网络北大实验室；深圳市信息论与未来网络重点实验室
//
package lf

import (
	"net"
)

type TcpTransport struct {
	StreamTransport
}

// Init
// @Description:  初始化 TcpTransport
// @receiver t
// @param conn
//
func (t *TcpTransport) Init(conn net.Conn) {
	t.conn = conn
	t.localAddr = conn.LocalAddr().String()
	t.localUri = "tcp://" + t.localAddr
	t.remoteAddr = conn.RemoteAddr().String()
	t.remoteUri = "tcp://" + t.remoteAddr
	t.recvBuf = make([]byte, 1024*1024*4)
	t.recvLen = 0
}
