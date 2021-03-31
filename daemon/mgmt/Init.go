package mgmt

import (
	"minlib/component"
	"minlib/mgmt"
	"minlib/packet"
)

// 初始化全局调度器
var dispatcher = CreateDispatcher()
var fibManager = CreateFibManager()
var csManager = CreateCsManager()
var faceManager = CreateFaceManager()

// 初始化函数
func init() {
	topPrefix, _ := component.CreateIdentifierByString("/min-mir/mgmt/localhost")
	dispatcher.AddTopPrefix(topPrefix)
	//注册add/delete/list命令
	fibManager.Init()
	csManager.Init()
	faceManager.Init()
}

func authorization(topPrefix *component.Identifier, interest *packet.Interest,
	parameters *mgmt.ControlParameters,
	accept AuthorizationAccept,
	reject AuthorizationReject) bool {

	err := parameters.Parse(interest)
	if err != nil {
		reject(err.(mgmt.ControlParametersError).Type)
		return false
	}
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
	// 验证签名不通过
	if err:=dispatcher.KeyChain.VerifyInterest(interest);err!=nil{

	}
	accept()
	return true
}

func authorizationAccept() {

}

func authorizationReject(errorType int) {

}
