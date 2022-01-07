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
// @Author: Jianming Que
// @Description:
// @Version: 1.0.0
// @Date: 2021/11/12 5:34 下午
// @Copyright: MIN-Group；国家重大科技基础设施——未来网络北大实验室；深圳市信息论与未来网络重点实验室
//
package table

import (
	"fmt"
	"github.com/bluele/gcache"
	"minlib/packet"
	"strings"
)

// UniversalCSPolicy 统一的缓存策略实现，基于gcache实现了LFU, LRU and ARC缓存替换策略
//
// @Description:
//
type UniversalCSPolicy struct {
	cache gcache.Cache
}

// NewUniversalCSPolicy 新建一个 UniversalCSPolicy
//
// @Description:
// @return *UniversalCSPolicy
//
func NewUniversalCSPolicy(capacity int, cacheType string) (*UniversalCSPolicy, error) {
	lruCSPolicy := new(UniversalCSPolicy)
	return lruCSPolicy, lruCSPolicy.Init(capacity, cacheType)
}

// Init 初始化 UniversalCSPolicy
//
// @Description:
// @receiver L
// @param capacity
//
func (L *UniversalCSPolicy) Init(capacity int, cacheType string) error {
	cacheBuilder := gcache.New(capacity)
	switch strings.ToLower(cacheType) {
	case "lru":
		cacheBuilder = cacheBuilder.LRU()
	case "lfu":
		cacheBuilder = cacheBuilder.LFU()
	case "arc":
		cacheBuilder = cacheBuilder.ARC()
	default:
		return UniversalCSPolicyError{
			msg: "Not support cache policy: " + cacheType + ", require: LRU, LFU, ARC",
		}
	}
	L.cache = cacheBuilder.
		Build()
	return nil
}

// Insert 缓存一个数据包
//
// @Description:
// @param data
// @return *CSEntry 返回缓存成功的CS条目
//
func (L *UniversalCSPolicy) Insert(data *packet.Data) (*CSEntry, error) {
	key := data.GetName().ToUri()
	if item, err := L.cache.Get(key); err != nil {
		// 不存在，则构建一个 CSEntry 插入
		csEntry := NewCSEntry(data)
		if err := L.cache.Set(key, csEntry); err != nil {
			return nil, err
		}
		return csEntry, nil
	} else {
		// 存在
		return item.(*CSEntry), nil
	}
}

// Find 根据传入的 Interest 查询CS表中是否缓存有与之匹配的 data
//
// @Description:
// 根据设计的不同，Interest和CS条目匹配的规则也可以有不同的定义：
//  1. NDN里面，Interest支持CanBePrefix字段，所以查询CS的时候不是精确匹配，是前缀匹配；
//  2. 在MIN当前的设计里面，为了加快并行转发并简化设计，可能会把Interest查询CS设计为hash查找；
//     (MIN里面也定义了CanBePrefix字段，但是暂时没有实现该效果，所以这个字段目前是无效字段）
// @param interest
// @return *CSEntry
//
func (L *UniversalCSPolicy) Find(interest *packet.Interest) (*CSEntry, error) {
	key := interest.GetName().ToUri()
	if item, err := L.cache.Get(key); err != nil {
		return nil, err
	} else {
		return item.(*CSEntry), nil
	}
}

// Size 返回已缓存的数据包的数量
//
// @Description:
// @return int
//
func (L *UniversalCSPolicy) Size() int {
	return L.cache.Len(false)
}

/////////////////////////////////////////////////////////////////////////////////////////////////////////
///// 错误处理
/////////////////////////////////////////////////////////////////////////////////////////////////////////

type UniversalCSPolicyError struct {
	msg string
}

func (u UniversalCSPolicyError) Error() string {
	return fmt.Sprintf("UniversalCSPolicyError: %s", u.msg)
}
