//
// @Author: weiguohua
// @Description:
// @Version: 1.0.0
// @Date: 2021/3/16 下午3:50
// @Copyright: MIN-Group；国家重大科技基础设施——未来网络北大实验室；深圳市信息论与未来网络重点实验室
//
package lf

import (
	"log"
	"net"
	"strconv"
)

//
// @Description:
//
type UdpListener struct {
	udpPort uint16
	conn    *net.UDPConn
}

func (u *UdpListener) Init() {
	u.udpPort = 13899
}

//
// @Description: 创建一个udp类型的logicFace
// @receiver t
// @param conn	新udp 句柄
//
func (u *UdpListener) createUdpLogicFace(conn *net.UDPConn) {
	logicFace, _ := createUdpLogicFace(conn)
	logicFace.Start()
}

//
// @Description:  启动监听协程
// @receiver t
//
func (u *UdpListener) Start() {
	udpAddr, err := net.ResolveUDPAddr("udp4", "0.0.0.0:"+strconv.Itoa(int(u.udpPort)))
	if err != nil {
		log.Fatal(err)
		return
	}
	conn, err := net.ListenUDP("udp", udpAddr)
	if err != nil {
		log.Fatal(err)
	}
	u.conn = conn
	u.createUdpLogicFace(conn)
}
