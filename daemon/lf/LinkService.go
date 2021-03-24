//
// @Author: weiguohua
// @Description:
// @Version: 1.0.0
// @Date: 2021/3/16 上午11:31
// @Copyright: MIN-Group；国家重大科技基础设施——未来网络北大实验室；深圳市信息论与未来网络重点实验室
//
package lf

import (
	"log"
	"minlib/encoding"
	"minlib/packet"
)

//
// @Description:  链路服务层，用于分包发送，把接收到的包分片合并
//		LinkService-LogicFace-Transport是一个一一对应的关系，他们相互绑定
//		在一个收包流程中网络数据最开始是通过transport流入的，由transport调用LinkService的 receive函数处理接收到的网络包，
//		再由linkService调用logicFace的receive函数。
//		在一个发送包的流程中，由logicFace调用linkService的发包函数，再由linkService调用transport的发包函数
//
type LinkService struct {
	transport    ITransport   // 传输通道
	logicFace    *LogicFace   // LinkService关联的logicFace
	lpReassemble LpReassemble // 包分片合并器

	mtu              int    // MTU大小
	lpPacketHeadSize int    // lpPacket 编码成数组时的头部大小
	lpPacketId       uint64 // 发送的lpPacket 编码
}

//
// @Description: 	计算lpPacket头部大小
// @receiver l
//
func (l *LinkService) calculateLpPacketHeadSize() {
	var lpPacket packet.LpPacket
	lpPacket.SetId(0)
	lpPacket.SetFragmentSeq(0)
	lpPacket.SetFragmentNum(2)
	var encoder encoding.Encoder
	err := encoder.EncoderReset(encoding.MaxPacketSize, 0)
	if err != nil {
		log.Fatal("cannot calculate lpPacketHeadSize in LinkService init", err)
	}
	l.lpPacketHeadSize, err = lpPacket.WireEncode(&encoder)
	if err != nil {
		log.Fatal("cannot calculate lpPacketHeadSize in LinkService init", err)
	}
}

//
// @Description: 初始化linkService
// @receiver l
// @param mtu	MTU，最大传输单元
//
func (l *LinkService) Init(mtu int) {
	l.mtu = mtu
	l.lpReassemble.Init()
	l.calculateLpPacketHeadSize()
	l.lpPacketId = 0
}

//
// @Description: 从lpPacket中提取出MINPacket对象
// @receiver l
// @param lpPacket  LpPacket 对象指针
// @return *packet.MINPacket	MINPacket对象指针
// @return error	提取失败错误信息
//
func (l *LinkService) getMINPacketFromLpPacket(lpPacket *packet.LpPacket) (*packet.MINPacket, error) {
	payload := lpPacket.GetValue()
	block, err := encoding.CreateBlockByBuffer(payload, true)
	if err != nil {
		return nil, err
	}
	var minPacket packet.MINPacket
	err = minPacket.WireDecode(block)
	if err != nil {
		return nil, err
	}
	return &minPacket, nil
}

//
// @Description: 	收到lpPacket包的处理函数，该函数被相关联的 transport 的 receive 函数调用
// @receiver l
// @param lpPacket 	lpPacket对象指针
//
func (l *LinkService) ReceivePacket(lpPacket *packet.LpPacket) {

	// 未分包，只有一个包
	if lpPacket.GetFragmentNum() == 1 {
		minPacket, err := l.getMINPacketFromLpPacket(lpPacket)
		if err != nil {
			log.Println(err)
			return
		}
		l.logicFace.ReceivePacket(minPacket)
		return
	}
	reassembleLpPacket := l.lpReassemble.ReceiveFragment(l.transport.GetRemoteUri(), lpPacket)
	if reassembleLpPacket == nil {
		return
	}
	minPacket, err := l.getMINPacketFromLpPacket(lpPacket)
	if err != nil {
		log.Println(err)
		return
	}
	l.logicFace.ReceivePacket(minPacket)
}

//
// @Description: 发送一个lp包分片
// @receiver l
// @param buf	分片的数据
// @param bufLen	数据长度
// @param fragmentId	分片号
// @param fragmentNum	分片数
// @param fragmentSeq	第几块分片，从0开始
//
func (l *LinkService) sendFragment(buf []byte, bufLen int, fragmentId, fragmentNum, fragmentSeq uint64) {
	var lpPacket packet.LpPacket
	lpPacket.SetId(fragmentId)
	lpPacket.SetFragmentNum(fragmentNum)
	lpPacket.SetFragmentSeq(fragmentSeq)
	lpPacket.SetValue(buf[:bufLen])
	l.transport.Send(&lpPacket)
}

//
// @Description: 发送指定长度的一段数据，如果数据过长，会调用sendFragment发送多个分片
// @receiver l
// @param buf	要发送的数据指针
// @param bufLen	数据长度
//
func (l *LinkService) sendByteBuffer(buf []byte, bufLen int) {
	fragmentLen := l.mtu - l.lpPacketHeadSize
	startIdx := 0
	fragmentSeq := 0
	fragmentNum := bufLen / fragmentLen
	if bufLen%fragmentLen != 0 {
		fragmentNum++
	}
	for startIdx < bufLen {
		if fragmentLen > bufLen-startIdx {
			fragmentLen = bufLen - startIdx
		}
		l.sendFragment(buf[startIdx:startIdx+fragmentLen], fragmentLen, l.lpPacketId, uint64(fragmentNum),
			uint64(fragmentSeq))
	}
	l.lpPacketId++
}

//
// @Description: 	发送一个兴趣包
// @receiver l
// @param interest	兴趣包对象指针
//
func (l *LinkService) SendInterest(interest *packet.Interest) {
	var encoder encoding.Encoder
	err := encoder.EncoderReset(encoding.MaxPacketSize, 0)
	if err != nil {
		log.Println(err)
		return
	}
	bufLen, err := interest.WireEncode(&encoder)
	if err != nil {
		log.Println(err)
		return
	}
	buf, err := encoder.GetBuffer()
	if err != nil {
		log.Println(err)
		return
	}
	l.sendByteBuffer(buf, bufLen)

}

//
// @Description:  发送一个数据包
// @receiver l
// @param data	数据包对象指针
//
func (l *LinkService) SendData(data *packet.Data) {
	var encoder encoding.Encoder
	err := encoder.EncoderReset(encoding.MaxPacketSize, 0)
	if err != nil {
		log.Println(err)
		return
	}
	bufLen, err := data.WireEncode(&encoder)
	if err != nil {
		log.Println(err)
		return
	}
	buf, err := encoder.GetBuffer()
	if err != nil {
		log.Println(err)
		return
	}
	l.sendByteBuffer(buf, bufLen)

}

//
// @Description:  发送一个Nack包
// @receiver l
// @param nack	Nack包对象指针
//
func (l *LinkService) SendNack(nack *packet.Nack) {
	var encoder encoding.Encoder
	err := encoder.EncoderReset(encoding.MaxPacketSize, 0)
	if err != nil {
		log.Println(err)
		return
	}
	bufLen, err := nack.WireEncode(&encoder)
	if err != nil {
		log.Println(err)
		return
	}
	buf, err := encoder.GetBuffer()
	if err != nil {
		log.Println(err)
		return
	}
	l.sendByteBuffer(buf, bufLen)

}

//
// @Description: 	发送一个普通推送式网络包
// @receiver l
// @param cPacket
//
func (l *LinkService) SendCPacket(cPacket *packet.CPacket) {
	var encoder encoding.Encoder
	err := encoder.EncoderReset(encoding.MaxPacketSize, 0)
	if err != nil {
		log.Println(err)
		return
	}
	bufLen, err := cPacket.WireEncode(&encoder)
	if err != nil {
		log.Println(err)
		return
	}
	buf, err := encoder.GetBuffer()
	if err != nil {
		log.Println(err)
		return
	}
	l.sendByteBuffer(buf, bufLen)
}
