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

// Package table
// @Author: yzy
// @Description:
// @Version: 1.0.0
// @Date: 2021/12/21 4:32 PM
// @Copyright: MIN-Group；国家重大科技基础设施——未来网络北大实验室；深圳市信息论与未来网络重点实验室
//

package table

import (
	"minlib/component"
	"minlib/packet"
	"sync"
	"time"
)

type CSEntry struct {
	data      *packet.Data     // 数据包指针
	StaleTime int64            // 不新鲜时间
	Interest  *packet.Interest // 兴趣包指针
	RWlock    *sync.RWMutex    // 读写锁
}

// NewCSEntry 获取表项中的数据包指针
func NewCSEntry(data *packet.Data) *CSEntry {
	var c = &CSEntry{}
	c.data = data
	c.StaleTime = time.Now().Unix()
	c.Interest = &packet.Interest{}
	c.RWlock = new(sync.RWMutex)
	return c
}

func (c *CSEntry) GetData() *packet.Data {
	return c.data
}

// GetIdentifier 获取表项中数据包的标识指针
func (c *CSEntry) GetIdentifier() *component.Identifier {
	return c.data.GetName()
}

// GetStaleTime 获得表项变旧时间
func (c *CSEntry) GetStaleTime() int64 {
	c.RWlock.RLock()
	defer c.RWlock.RUnlock()
	return c.StaleTime
}

// IsStale 判断表项是否已经变得不新鲜
func (c *CSEntry) IsStale() bool {
	c.RWlock.RLock()
	defer c.RWlock.RUnlock()
	return c.StaleTime < time.Now().Unix()
}

// UpdateStaleTime 更新表项的变旧时间
func (c *CSEntry) UpdateStaleTime(newStaleTime int64) {
	c.RWlock.Lock()
	defer c.RWlock.Unlock()
	c.StaleTime = newStaleTime
}

// CanSatisfy 判断表项是否可以与某个兴趣包匹配 参考C++语言代码
func (c *CSEntry) CanSatisfy(interest *packet.Interest) bool {
	if !interest.MatchesData(c.data) {
		return false
	}
	if interest.GetMustBeRefresh() == true && c.IsStale() {
		return false
	}
	return true
}
