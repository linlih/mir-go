//
// @Author: Jianming Que
// @Description:
// @Version: 1.0.0
// @Date: 2021/3/10 11:37 上午
// @Copyright: MIN-Group；国家重大科技基础设施——未来网络北大实验室；深圳市信息论与未来网络重点实验室
//
package mgmt

import (
	"minlib/component"
	"minlib/encoding"
	"minlib/mgmt"
)

//
// 数据集上下文，由 Dispatcher 创建，传递给具体的管理模块，管理模块调用本上下文对象将数据通过 Dispatcher 发出
//
// @Description:
//
type StatusDatasetContext struct {
	Prefix    component.Identifier // 要发布的数据的前缀
	FreshTime int                  // 生成的 Data 的新鲜期，默认为 1 s
}

//
// 添加一个要发送的 Block 作为响应
//
// @Description:
// @receiver s
// @param block
//
func (s *StatusDatasetContext) Append(block *encoding.Block) {

}

//
// 所有数据 Append 完毕之后，调用本方法，会生成一个可以标识数据集结束的 Data ，用户侧收到这种特殊的包即可判定本次数据拉取结束
//
// @Description:
// @receiver s
// @param block
//
func (s *StatusDatasetContext) Finish() {

}

//
// 如果在生成数据集的过程中发送了错误，可以通过本方法往用户侧发送一个表示错误的响应
//
// @Description:
// @receiver s
// @param response
//
func (s *StatusDatasetContext) Reject(response mgmt.ControlResponse) {

}
