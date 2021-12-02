// Package table
// @Author: Jianming Que
// @Description:
// @Version: 1.0.0
// @Date: 2021/11/12 11:36 上午
// @Copyright: MIN-Group；国家重大科技基础设施——未来网络北大实验室；深圳市信息论与未来网络重点实验室
//
package table

import (
	"minlib/packet"
	"mir-go/daemon/common"
)

// UniversalCS 基于Hash表实现的 ContentStore
//
// @Description:
//
type UniversalCS struct {
	csPolicy ICSPolicy
}

// NewUniversalCS 新建一个 UniversalCS
//
// @Description:
// @param config
// @return *UniversalCS
//
func NewUniversalCS(config *common.MIRConfig) (*UniversalCS, error) {
	hashCS := new(UniversalCS)
	return hashCS, hashCS.Init(config)
}

// Init 初始化 UniversalCS
//
// @Description:
// @receiver h
// @param config
// @return error
//
func (h *UniversalCS) Init(config *common.MIRConfig) error {
	if policy, err := NewUniversalCSPolicy(config.TableConfig.CSSize, config.TableConfig.CSReplaceStrategy); err != nil {
		return err
	} else {
		h.csPolicy = policy
	}
	return nil
}

// Size 返回已缓存的数据包的数量
//
// @Description:
// @receiver h
// @return int
//
func (h *UniversalCS) Size() int {
	return h.csPolicy.Size()
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
func (h *UniversalCS) Find(interest *packet.Interest) (*CSEntry, error) {
	return h.csPolicy.Find(interest)
}

// Insert 将传入的 data 缓存到CS当中
//
// @Description:
// 插入过程需要根据CS自己定义的缓存替换策略，来替换、踢出或者更新CS条目
// @param data
// @return *CSEntry
//
func (h *UniversalCS) Insert(data *packet.Data) (*CSEntry, error) {
	return h.csPolicy.Insert(data)
}
