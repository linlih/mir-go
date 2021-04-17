// Package lf
// @Author: weiguohua
// @Description:
// @Version: 1.0.0
// @Date: 2021/3/16 上午11:35
// @Copyright: MIN-Group；国家重大科技基础设施——未来网络北大实验室；深圳市信息论与未来网络重点实验室
//
package lf

import "net"

type UnixStreamTransport struct {
	StreamTransport
}

func (u *UnixStreamTransport) Init(conn net.Conn) {
	u.conn = conn
	u.localAddr = conn.LocalAddr().String()
	u.localUri = "unix://" + u.localAddr
	u.remoteAddr = conn.RemoteAddr().String()
	u.remoteUri = "unix://" + u.remoteAddr
	u.recvBuf = make([]byte, 1024*1024*4)
	u.recvLen = 0
}
