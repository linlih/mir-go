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
// @Date: 2021/11/1 10:33 上午
// @Copyright: MIN-Group；国家重大科技基础设施——未来网络北大实验室；深圳市信息论与未来网络重点实验室
//
package table

import "minlib/packet"

// ICS 定义CS（ContentStore）表的通用行为，每一个CS表的实现都应该实现本接口
//
// @Description:
//
type ICS interface {
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
	Find(interest *packet.Interest) (*CSEntry, error)

	// Insert 将传入的 data 缓存到CS当中
	//
	// @Description:
	// 插入过程需要根据CS自己定义的缓存替换策略，来替换、踢出或者更新CS条目
	// @param data
	// @return *CSEntry
	//
	Insert(data *packet.Data) (*CSEntry, error)

	// Size 返回已缓存的数据包的数量
	//
	// @Description:
	// @return int
	//
	Size() int
}
