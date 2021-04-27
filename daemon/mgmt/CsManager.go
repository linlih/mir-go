// Package mgmt
// @Author: yzy
// @Description:
// @Version: 1.0.0
// @Date: 2021/3/10 3:13 下午
// @Copyright: MIN-Group；国家重大科技基础设施——未来网络北大实验室；深圳市信息论与未来网络重点实验室
//
package mgmt

import (
	"minlib/common"
	"minlib/component"
	"minlib/mgmt"
	"minlib/packet"
	"mir-go/daemon/lf"
	"mir-go/daemon/table"
)

// CsManager
// CS管理模块结构体
//
// @Description:CS管理模块结构体
//
type CsManager struct {
	cs             *table.CS // CS表
	logicFaceTable *lf.LogicFaceTable
	enableServe    bool // 是否可以展示信息
	enableAdd      bool // 是否可以添加缓存
}

// CreateCsManager
// 创建CS管理模块函数
//
// @Description:创建管理模块,进行初始化并返回指针
// @Return:*CsManager
//
func CreateCsManager() *CsManager {
	return &CsManager{
		cs:          new(table.CS),
		enableServe: true,
		enableAdd:   true,
	}
}

// Init
// CS管理模块初始化注册行为函数
//
// @Description:对CS管理模块注册三个必须的函数add、delete、list
// @receiver c
//
func (c *CsManager) Init(dispatcher *Dispatcher, logicFaceTable *lf.LogicFaceTable) {
	c.logicFaceTable = logicFaceTable
	identifier, _ := component.CreateIdentifierByString("/cs-mgmt/delete")
	err := dispatcher.AddControlCommand(identifier, dispatcher.authorization, c.ValidateParameters, c.changeConfig)
	if err != nil {
		common.LogError("cs add delete-command fail,the err is:", err)
	}
	identifier, _ = component.CreateIdentifierByString("/cs-mgmt/list")
	err = dispatcher.AddStatusDataset(
		identifier,
		dispatcher.authorization,
		func(parameters *component.ControlParameters) bool {
			return true
		},
		c.serveInfo,
	)
	if err != nil {
		common.LogError("cs add list-command fail,the err is:", err)
	}
}

// TODO:后续进行实现，配置CS表读写权限等
//
// 修改配置函数
//
// @Description:对CS管理模块配置进行修改，如是否可读，是否可以插入新的数据
// @receiver c
//
func (c *CsManager) changeConfig(topPrefix *component.Identifier, interest *packet.Interest,
	parameters *component.ControlParameters) *mgmt.ControlResponse {
	c.enableServe = true
	return nil
}

//
// 获取CS管理模块的服务信息
//
// @Description:获取CS管理模块的服务信息，分片发送给客户端，信息包括配置信息、条目数量、命中缓存次数等
// @receiver c
//
func (c *CsManager) serveInfo(topPrefix *component.Identifier, interest *packet.Interest,
	parameters *component.ControlParameters,
	context *StatusDatasetContext) {
	//var response *mgmt.ControlResponse
	//if c.enableServe {
	//	var CSInfo = struct {
	//		enableServe bool
	//		enableAdd   bool
	//		size        uint64
	//		hits        uint64
	//		misses      uint64
	//	}{
	//		enableServe: c.enableServe,
	//		enableAdd:   c.enableAdd,
	//		size:        c.cs.Size(),
	//		hits:        c.cs.Hits,
	//		misses:      c.cs.Misses,
	//	}
	//	data, err := json.Marshal(CSInfo)
	//	if err != nil {
	//		response = MakeControlResponse(400, "mashal CSInfo fail , the err is:"+err.Error(), "")
	//		context.responseSender(response, interest)
	//	}
	//	context.data = data
	//	dataList := context.Append()
	//	if dataList == nil {
	//		response = MakeControlResponse(400, "slice data packet err!", "")
	//		context.responseSender(response, interest)
	//		return
	//	} else {
	//		for _, data := range dataList {
	//			// 包编码放在dataSender中
	//			context.dataSender(data)
	//		}
	//	}
	//} else {
	//	response = MakeControlResponse(400, "have no Permission to get CsInfo!", "")
	//	context.responseSender(response, interest)
	//}
}

// ValidateParameters
// 参数验证函数
//
// @Description:对传入的控制参数进行参数验证，条件判断语句中的为必需字段，有一项不存在则错误
// @receiver c
// @Return:bool
//
func (c *CsManager) ValidateParameters(parameters *component.ControlParameters) bool {
	if parameters.ControlParameterPrefix.IsInitial() &&
		parameters.ControlParameterCount.IsInitial() &&
		parameters.ControlParameterCapacity.IsInitial() {
		return true
	}
	return false
}
