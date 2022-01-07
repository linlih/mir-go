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
// @Date: 2021/11/12 5:33 下午
// @Copyright: MIN-Group；国家重大科技基础设施——未来网络北大实验室；深圳市信息论与未来网络重点实验室
//
package table

import "minlib/packet"

// ICSPolicy ContentStore 缓存替换策略接口，定义了CS缓存替换策略的通用行为，所有的CS缓存替换策略都需要实现本接口
//
// @Description:
//
type ICSPolicy interface {

	// Insert 缓存一个数据包
	//
	// @Description:
	// @param data
	// @return *CSEntry 返回缓存成功的CS条目
	//
	Insert(data *packet.Data) (*CSEntry, error)

	// Find 根据兴趣包查找匹配的数据包
	//
	// @Description:
	// @param interest
	// @return *CSEntry
	//
	Find(interest *packet.Interest) (*CSEntry, error)

	// Size 返回已缓存的数据包的数量
	//
	// @Description:
	// @return int
	//
	Size() int
}
