//
// @Author: yzy
// @Description:
// @Version: 1.0.0
// @Date: 2021/3/10 3:13 下午
// @Copyright: MIN-Group；国家重大科技基础设施——未来网络北大实验室；深圳市信息论与未来网络重点实验室
//
package mgmt

import (
	"minlib/component"
	"minlib/mgmt"
	"minlib/packet"
	"mir-go/daemon/common"
)

// 初始化全局调度器
var dispatcher = CreateDispatcher()
var fibManager = CreateFibManager()
var csManager = CreateCsManager()
var faceManager = CreateFaceManager()

//
// mgmt包初始化函数，在引入包的时候自动执行
//
// @Description:mgmt包初始化函数，注册本地管理顶级前缀，对所有的表管理模块初始绑定函数
//
func init() {
	topPrefix, _ := component.CreateIdentifierByString("/min-mir/mgmt/localhost")
	dispatcher.AddTopPrefix(topPrefix)
	//注册add/delete/list命令
	fibManager.Init()
	csManager.Init()
	faceManager.Init()
}

//
// 授权验证函数
//
// @Description:对收到的兴趣包中的参数进行解析，并验证权限
// @Return:bool
//
func authorization(topPrefix *component.Identifier, interest *packet.Interest,
	parameters *mgmt.ControlParameters,
	accept AuthorizationAccept,
	reject AuthorizationReject) bool {
	if _, ok := dispatcher.topPrefixList[topPrefix.ToUri()]; !ok {
		// 顶级域不存在
		reject(5)
		return false
	}
	// 没有权限
	if topPrefix.ToUri() == "" {
		reject(6)
		return false
	}
	if err := parameters.Parse(interest); err != nil {
		common.LogError("解析控制参数错误！the err is:", err)
	}
	// 验证签名不通过
	if err := dispatcher.KeyChain.VerifyInterest(interest); err != nil {

	}
	accept()
	return true
}

// TODO:暂未实现
// 授权成功回调
//
// @Description:如果授权成功执行此函数进行相应处理
//
func authorizationAccept() {

}

// TODO:暂未实现
// 授权失败回调
//
// @Description:如果授权失败执行此函数进行相应处理
//
func authorizationReject(errorType int) {

}
