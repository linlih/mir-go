//
// @Author: weiguohua
// @Description:
// @Version: 1.0.0
// @Date: 2021/3/22 下午3:24
// @Copyright: MIN-Group；国家重大科技基础设施——未来网络北大实验室；深圳市信息论与未来网络重点实验室
//
package lf

import (
	"errors"
	"log"
	"minlib/encoding"
	"minlib/packet"
	"net"
)

//
// @Description: 流式通道共用类， TcpTransport, UnixStreamTransport
//
type StreamTransport struct {
	Transport
	conn    net.Conn
	recvBuf []byte
	recvLen uint64
}

//
// @Description:
// @receiver t
//
func (t *StreamTransport) Close() {
	err := t.conn.Close()
	if err != nil {
		log.Println(err)
	}
}

//
// @Description:
// @receiver t
// @param lpPacket
//
func (t *StreamTransport) Send(lpPacket *packet.LpPacket) {
	encodeBufLen, encodeBuf := t.encodeLpPacket2ByteArray(lpPacket)
	if encodeBufLen < 0 {
		return
	}
	writeLen := 0
	for writeLen < encodeBufLen {
		writeRet, err := t.conn.Write(encodeBuf[:encodeBufLen])
		if err != nil {
			log.Println("send to tcp transport error")
			// TODO close the face
			return
		}
		writeLen += writeRet
	}
}

//
// @Description: 从接收缓冲区中读取包并调用linkService的ReceivePacket函数处理包
// @receiver t
// @param buf	接收缓冲区分片
// @return error	错误信息
// @return uint64	读取包的长度
//
func (t *StreamTransport) readPktAndDeal(buf []byte, bufLen uint64) (error, uint64) {
	// 如果接收到的数据长度小于 LpPacket type 字段的长度 3字节 则要等待
	if bufLen < uint64(encoding.SizeOfVarNumber(encoding.TlvLpPacket)) {
		return nil, 0
	}
	pktType, err := encoding.ReadVarNumber(buf, 0)
	if err != nil {
		log.Println(err)
		return err, 0
	}
	// 如果数据类型的 TLV 和 type值不等于   encoding.TlvLpPacket， 则接收出错，应该关闭当前logicFace
	if pktType != encoding.TlvLpPacket {
		log.Println("receive error pkt")
		return errors.New("receive lpPacket from tcp type error"), 0
	}
	// 如果接收到的数据长度小于 LpPacket 的小于长度 则要等待
	if bufLen < uint64(t.linkService.lpPacketHeadSize) {
		return nil, 0
	}
	pktTypeLen := encoding.SizeOfVarNumber(pktType)
	pktLen, err := encoding.ReadVarNumber(buf, encoding.VlInt(pktTypeLen))
	totalPktLen := uint64(pktTypeLen) + uint64(encoding.SizeOfVarNumber(pktLen)) + uint64(pktLen)
	if bufLen >= totalPktLen {
		lpPacket, err := t.parseByteArray2LpPacket(buf[:totalPktLen])
		if err != nil {
			log.Println("parse lpPacket error")
		} else {
			t.linkService.ReceivePacket(lpPacket)
		}
		return nil, totalPktLen
	}
	return nil, 0
}

//
// @Description: 接收到数据后，处理包
// @receiver t
// @return error	如果处理包出错，则返回错误信息
//
func (t *StreamTransport) onReceive() error {
	err, pktLen := t.readPktAndDeal(t.recvBuf[:t.recvLen], t.recvLen)
	if err != nil {
		return err
	}
	var dealLen uint64 = 0
	// 循环多次尝试从接收缓冲区中读出包并处理
	for err == nil && pktLen > 0 {
		dealLen += pktLen
		err, pktLen = t.readPktAndDeal(t.recvBuf[dealLen:t.recvLen-dealLen], t.recvLen-dealLen)
	}
	if err != nil {
		return err
	}
	if dealLen > 0 {
		copy(t.recvBuf[:], t.recvBuf[dealLen:t.recvLen])
		t.recvLen -= dealLen
	}
	return nil
}

//
// @Description:  用协程调用，不断地从Tcp通道中读出数据
// @receiver t
//
func (t *StreamTransport) Receive() {
	for true {
		recvRet, err := t.conn.Read(t.recvBuf[t.recvLen:])
		if err != nil {
			log.Println("recv from tcp transport error")
			// TODO close the face
		}
		t.recvLen += uint64(recvRet)
		err = t.onReceive()
		if err != nil {
			log.Println("recv from tcp transport error")
			// TODO close the face
		}
	}
}

//
// @Description:
// @receiver t
// @return string
//
func (t *StreamTransport) GetRemoteUri() string {
	return t.remoteUri
}

//
// @Description:
// @receiver t
// @return string
//
func (t *StreamTransport) GetLocalUri() string {
	return t.localUri
}
