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
