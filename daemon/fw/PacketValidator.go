//
// @Author: Jianming Que
// @Description:
// @Version: 1.0.0
// @Date: 2021/4/1 3:20 下午
// @Copyright: MIN-Group；国家重大科技基础设施——未来网络北大实验室；深圳市信息论与未来网络重点实验室
//
package fw

import (
	"github.com/panjf2000/ants"
	"minlib/security"
	"mir-go/daemon/common"
	"mir-go/daemon/lf"
)

//
// 表示一个包验证器，本验证器会并发的对收到的网络包进行签名验证，并且在
//
// @Description:
//
type PacketValidator struct {
	_pool        *ants.Pool         // 协程池，用于并发验签
	packetQueue  *BlockQueue        // 一个阻塞队列，用于和 Forwarder 进行通信
	keyChain     *security.KeyChain // 一个KeyChain，用于包签名验证
	cap          int                // 协程池容量
	needValidate bool               // 是否需要进行验证（如果不开启签名验证，则直接传递给缓存队列即可，无需开启线程池）
}

//
// 初始化包验证器
//
// @Description:
// @receiver p
// @param cap					协程池的大小
// @param needValidate			是否需要开启签名验证
// @param packetQueue			与 Forwarder 共同持有的一个阻塞队列
//
func (p *PacketValidator) Init(cap int, needValidate bool, packetQueue *BlockQueue) {
	p.cap = cap
	p.packetQueue = packetQueue
	// 当且仅当需要进行签名验证时，才开启协程池
	if needValidate {
		if keyChain, err := security.CreateKeyChain(); err != nil {
			common.LogFatal("Create KeyChain failed! msg => %s", err.Error())
		} else {
			p.keyChain = keyChain
		}
		p._pool, _ = ants.NewPool(cap)
		if err := p.keyChain.InitialKeyChain(); err != nil {
			// 如果初始化KeyChain失败，则认为是严重错误直接抛出错误退出程序
			common.LogFatal("PacketValidator init KeyChain failed！ msg => %s", err.Error())
		}
	}
}

//
// 收到一个MINPacket
//
// @Description:
//	1. 如果开启了签名验证，则将收到的网络包交给协程池进行并发的验证，验证通过则放入 p.packetQueue
//	2. 如果没有开启签名验证，则直接将收到的网络包放入 p.packetQueue
// @receiver p
// @param data
//
func (p *PacketValidator) ReceiveMINPacket(data *lf.IncomingPacketData) {
	if !p.needValidate {
		// 如果不需要进行包验证，则直接放到队列中
		p.packetQueue.write(data)
		return
	}

	// 如果开启了包验证，则放到协程池里进行并发验证
	if err := p._pool.Submit(func() {
		// TODO: 这边需要检查一下 KeyChain 的签名验证方法是不是多线程安全的
		if err := p.keyChain.Verify(data.MinPacket); err == nil {
			// 验证成功
			common.LogDebugWithFields(data.ToFields(), "Verify Packet Success")
			// 验证成功之后将包放入队列中
			p.packetQueue.write(data)
		} else {
			// 验证失败
			common.LogDebugWithFields(data.ToFields(), "Verify Packet Failed")
		}
	}); err != nil {
		// 任务提交失败，输出错误
		common.LogError("PacketValidator create a packet verify task failed: %s", err.Error())
	}
}

//
// 关闭包验证器
//
// @Description:
// @receiver p
//
func (p *PacketValidator) Close() {
	if p._pool != nil {
		// 关闭协程池
		p._pool.Release()
	}
}
