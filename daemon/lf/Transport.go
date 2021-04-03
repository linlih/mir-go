//
// @Author: weiguohua
// @Description:
// @Version: 1.0.0
// @Date: 2021/3/16 上午11:32
// @Copyright: MIN-Group；国家重大科技基础设施——未来网络北大实验室；深圳市信息论与未来网络重点实验室
//
package lf

import (
	"minlib/encoding"
	"minlib/packet"
	"mir-go/daemon/common"
)

//
// @Description:  Tranport共用类
//
type Transport struct {
	localAddr   string
	remoteAddr  string
	localUri    string
	remoteUri   string
	linkService *LinkService
}

//
// @Description: 从[]byte中解析出LpPacket
// @receiver e
// @param pkt
// @return *packet.LpPacket	解析出的包
// @return error		解析失败错误
//
func parseByteArray2LpPacket(buf []byte) (*packet.LpPacket, error) {
	block, err := encoding.CreateBlockByBuffer(buf, true)
	if err != nil {
		common.LogWarn(err)
		return nil, err
	}
	if !block.IsValid() {
		common.LogWarn("recv packet from face invalid")
		return nil, err
	}
	var lpPacket packet.LpPacket
	err = lpPacket.WireDecode(block)
	if err != nil {
		common.LogWarn("parse to lpPacket error")
		return nil, err
	}
	return &lpPacket, nil
}

//
// @Description: 	将lpPacket编码成byte数组
// @receiver t
// @param lpPacket
// @return int     编码后byte数组的长度
// @return []byte	编码得到的byte数组
//
func encodeLpPacket2ByteArray(lpPacket *packet.LpPacket) (int, []byte) {
	var encoder encoding.Encoder
	err := encoder.EncoderReset(encoding.MaxPacketSize, 0)
	encodeBufLen, err := lpPacket.WireEncode(&encoder)
	if err != nil {
		common.LogWarn(err)
		return -1, nil
	}
	encodeBuf, err := encoder.GetBuffer()
	if err != nil {
		common.LogWarn(err)
		return -1, nil
	}
	return encodeBufLen, encodeBuf
}

//
// @Description:
// @receiver t
// @return string
//
func (t *Transport) GetRemoteUri() string {
	return t.remoteUri
}

//
// @Description:
// @receiver t
// @return string
//
func (t *Transport) GetLocalUri() string {
	return t.localUri
}

//
// @Description:
// @receiver t
// @return string
//
func (t *Transport) GetRemoteAddr() string {
	return t.remoteAddr
}

//
// @Description:
// @receiver t
// @return string
//
func (t *Transport) GetLocalAddr() string {
	return t.localAddr
}
