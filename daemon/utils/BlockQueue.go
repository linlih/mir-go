// Package utils
// @Author: Jianming Que
// @Description:
// @Version: 1.0.0
// @Date: 2021/3/26 9:59 上午
// @Copyright: MIN-Group；国家重大科技基础设施——未来网络北大实验室；深圳市信息论与未来网络重点实验室
//
package utils

import (
	"fmt"
	"time"
)

// BlockQueue
// 基于 Chan 实现一个阻塞队列，用于 LogicFace 和 Forwarder 进行 FIFO 的包传递
//
// @Description:
//
type BlockQueue struct {
	bufChan chan interface{} // 一个用于传输 MINPacket 的通道，实现 FIFO 效果
	size    uint             // buffer size
}

// CreateBlockQueue
// 新建一个阻塞队列
//
// @Description:
// @param size		队列的容量
// @return *BlockQueue
//
func CreateBlockQueue(size uint) *BlockQueue {
	return &BlockQueue{
		bufChan: make(chan interface{}, size),
		size:    size,
	}
}

//
// 阻塞地从缓存队列里面读取一个数据
//
// @Description:
//	本操作会持续阻塞等待，如果队列里面一直为空，则会一直阻塞
// @receiver b
// @return interface{}
// @return error
//
func (b *BlockQueue) Read() interface{} {
	return <-b.bufChan
}

// ReadUntil
// 尝试从缓存队列里面读取一个数据，并且在队列为空时等待一段时间
//
// @Description:
//	1. 如果队列里面有数据，则立即返回
//	2. 如果队列里面为空吗，但是在等待期间，队列里面有了新数据数据，则返回一条数据
//	3. 如果队列里面为空，且等待了 waitTime ms 之后仍然为空，则返回错误
// @receiver b
// @param waitTime					队列为空时最长等待的时间
// @return interface{}
// @return error
//
func (b *BlockQueue) ReadUntil(waitTime uint) (interface{}, error) {
	ticker := time.NewTicker(time.Duration(waitTime) * time.Millisecond)
	select {
	case data := <-b.bufChan:
		return data, nil
	case <-ticker.C:
		return nil, timeoutError
	}
}

//
// 阻塞地往缓存队列写入一个 MINPacket
//
// @Description:
// @receiver b
// @param data
//
func (b *BlockQueue) Write(data interface{}) {
	b.bufChan <- data
}

// WriteUtil
// 尝试往缓存队列里面写入一个数据，并且在队列满时等待一段时间
//
// @Description:
//	1. 如果队列没有满，则写入数据并返回
//	2. 如果队列已满，但是在等待期间，队列里面有了新的空位，则写入数据并返回
//	3. 如果队列已满，且等待了 waitTime ms 之后仍然满，则返回错误
// @receiver b
// @param waitTime
// @param data
// @return error
//
func (b *BlockQueue) WriteUtil(waitTime uint, data interface{}) error {
	ticker := time.NewTicker(time.Duration(waitTime) * time.Millisecond)
	select {
	case b.bufChan <- data:
		return nil
	case <-ticker.C:
		return timeoutError
	}
}

/////////////////////////////////////////////////////////////////////////////////////////////////////////
///// 错误处理
/////////////////////////////////////////////////////////////////////////////////////////////////////////

type BlockQueueError struct {
	msg string
}

var (
	timeoutError = BlockQueueError{msg: fmt.Sprintf("Timeout for Read or Write!")}
)

func (b BlockQueueError) Error() string {
	return fmt.Sprintf("BlockQueueError: %s", b.msg)
}
