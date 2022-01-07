// Copyright [2022] [MIN-Group -- Peking University Shenzhen Graduate School Multi-Identifier Network Development Group]
//
// Licensed under the Apache License, Version 2.0 (the "License"): you may
// not use this file except in compliance with the License. You may obtain
// a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS, WITHOUT
// WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the
// License for the specific language governing permissions and limitations
// under the License.

// Package lf
// @Author: weiguohua
// @Description:
// @Version: 1.0.0
// @Date: 2021/3/18 上午10:36
// @Copyright: MIN-Group；国家重大科技基础设施——未来网络北大实验室；深圳市信息论与未来网络重点实验室
//
package lf

import (
	common2 "minlib/common"
	"minlib/packet"
)

const MaxFragmentNum = 1000

// PartialPacket
// @Description: 包重组哈希表的表项，主要用于保存接收到的包分片和记录超时时间
//
type PartialPacket struct {
	fragments          []*packet.LpPacket // 包分片数组
	fragCount          uint64             // 包分片数
	nReceivedFragments uint64             // 已接收到的包分片数
	dropTime           int64              // 重组超时时间点
}

// AddLpPacket
// @Description: 往包重组表中添加一个lpPacket
// @receiver p
// @param lpPacket
// @param curTime
//
func (p *PartialPacket) AddLpPacket(lpPacket *packet.LpPacket, curTime int64) {
	if lpPacket.GetFragmentNum() > MaxFragmentNum {
		common2.LogWarn("exceed max fragment number")
		return
	}
	if len(p.fragments) <= 0 { // 添加第一个包
		p.fragments = make([]*packet.LpPacket, lpPacket.GetFragmentNum())
		p.fragments[lpPacket.GetFragmentSeq()] = lpPacket
		p.dropTime = curTime + ReassembleTimeout
		p.nReceivedFragments = 1
		p.fragCount = lpPacket.GetFragmentNum()
		return
	}
	if p.fragCount <= lpPacket.GetFragmentSeq() || p.fragments[lpPacket.GetFragmentSeq()] != nil {
		return
	}
	p.fragments[lpPacket.GetFragmentSeq()] = lpPacket
	p.dropTime = curTime + ReassembleTimeout
	p.nReceivedFragments++
}

// DoReassemble
// @Description: 对已收到的包分片进行重组
// @receiver p
// @return *packet.LpPacket
//
func (p *PartialPacket) DoReassemble() *packet.LpPacket {
	if p.nReceivedFragments < p.fragCount {
		return nil
	}
	var buf []byte
	for _, e := range p.fragments {
		buf = append(buf, e.GetValue()...)
	}
	var lpPacket packet.LpPacket
	lpPacket.SetValue(buf)
	lpPacket.SetFragmentNum(1)
	lpPacket.SetFragmentSeq(0)
	lpPacket.SetId(0)
	return &lpPacket
}

// TimeoutEvent
// @Description: 超时事件
//
type TimeoutEvent struct {
	key         string // 重组器哈希表的key
	timeoutTime int64  // 超时时间
}

// TimeoutEventHeap
// @Description: 超时事件堆
//
type TimeoutEventHeap []TimeoutEvent

func (p *TimeoutEventHeap) Less(i, j int) bool {
	return (*p)[i].timeoutTime < (*p)[j].timeoutTime
}

func (p *TimeoutEventHeap) Len() int {
	return len(*p)
}

func (p *TimeoutEventHeap) Swap(i, j int) {
	(*p)[i], (*p)[j] = (*p)[j], (*p)[i]
}

func (p *TimeoutEventHeap) Push(val interface{}) {
	*p = append(*p, val.(TimeoutEvent))
}

func (p *TimeoutEventHeap) Pop() interface{} {
	old := *p
	*p = old[:len(old)-1]
	return old[len(old)-1]
}
