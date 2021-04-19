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
	"strconv"
)

// FibManager
// fib管理模块结构体
//
// @Description:fib管理模块结构体
//
type FibManager struct {
	fib            *table.FIB //fib表
	logicFaceTable *lf.LogicFaceTable
}

// CreateFibManager
// 创建fib管理模块函数
//
// @Description:创建fib管理模块函数并返回指针
//
func CreateFibManager() *FibManager {
	return &FibManager{
		fib: table.CreateFIB(),
	}
}

// Init
// fib管理模块初始化注册命令函数
//
// @Description:fib管理模块初始化注册命令函数
// @receiver f
//
func (f *FibManager) Init(dispatcher *Dispatcher, logicFaceTable *lf.LogicFaceTable) {
	f.logicFaceTable = logicFaceTable
	identifier, _ := component.CreateIdentifierByString("/" + mgmt.ManagementModuleFibMgmt + "/" + mgmt.FibManagementActionAdd)
	err := dispatcher.AddControlCommand(identifier, dispatcher.authorization, func(parameters *component.ControlParameters) bool {
		if parameters.ControlParameterPrefix.IsInitial() &&
			parameters.ControlParameterLogicFaceId.IsInitial() &&
			parameters.ControlParameterCost.IsInitial() {
			return true
		}
		return false
	}, f.AddNextHop)
	if err != nil {
		common.LogError("add add-command fail,the err is:", err)
	}
	identifier, _ = component.CreateIdentifierByString("/" + mgmt.ManagementModuleFibMgmt + "/" + mgmt.FibManagementActionDel)
	err = dispatcher.AddControlCommand(identifier, dispatcher.authorization, func(parameters *component.ControlParameters) bool {
		if parameters.ControlParameterPrefix.IsInitial() &&
			parameters.ControlParameterLogicFaceId.IsInitial() {
			return true
		}
		return false
	}, f.RemoveNextHop)
	if err != nil {
		common.LogError("add delete-command fail,the err is:", err)
	}
	identifier, _ = component.CreateIdentifierByString("/" + mgmt.ManagementModuleFibMgmt + "/" + mgmt.FibManagementActionList)
	err = dispatcher.AddStatusDataset(identifier, dispatcher.authorization, f.ListEntries)
	if err != nil {
		common.LogError("add list-command fail,the err is:", err)
	}

	// TODO: 加一个 registerIdentifier => register
	// 参数：prefix: Identifier，cost
	// 效果：在 FIB 添加一个条目，
}

// AddNextHop
// fib表中添加下一跳
//
// @Description:fib表中添加下一跳并返回添加结果
// @receiver f
//
func (f *FibManager) AddNextHop(topPrefix *component.Identifier, interest *packet.Interest,
	parameters *component.ControlParameters) *mgmt.ControlResponse {

	// 提取参数
	prefix := parameters.ControlParameterPrefix.Prefix()
	logicFaceId := parameters.ControlParameterLogicFaceId.LogicFaceId()
	cost := parameters.ControlParameterCost.Cost()

	// 标识前缀 不能太长 太长返回错误信息
	if prefix.Size() > table.MAX_DEPTH {
		common.LogError("add next hop fail,the err is:the prefix is too long")
		// 返回前缀太长的错误信息
		return MakeControlResponse(414, "the prefix is too long ,cannot exceed "+strconv.Itoa(table.MAX_DEPTH)+"components", "")
	}

	// 根据Id从table中取出 LogicFace
	face := f.logicFaceTable.GetLogicFacePtrById(logicFaceId)
	if face == nil {
		common.LogError("add next hop fail,the err is:", prefix.ToUri()+" logicFaceId:"+strconv.FormatUint(logicFaceId, 10)+"failed!")
		return MakeControlResponse(414, "the face is not found", "")
	}
	f.fib.AddOrUpdate(prefix, face, cost)
	return MakeControlResponse(200, "add next hop success", "")
}

// RemoveNextHop
// 根据logicface在fib表中删除下一跳
//
// @Description:根据logicface在fib表中删除下一跳并返回删除结果
// @receiver f
//
func (f *FibManager) RemoveNextHop(topPrefix *component.Identifier, interest *packet.Interest,
	parameters *component.ControlParameters) *mgmt.ControlResponse {
	// 获取前缀
	prefix := parameters.ControlParameterPrefix.Prefix()
	logicfaceId := parameters.ControlParameterLogicFaceId.LogicFaceId()
	// 根据Id从table中取出 logicface
	face := f.logicFaceTable.GetLogicFacePtrById(logicfaceId)
	if face == nil {
		common.LogError("add next hop fail,the err is:face is not found")
		return MakeControlResponse(410, "the face is not found", "")

	}
	fibEntry := f.fib.FindExactMatch(prefix)
	if fibEntry == nil {
		common.LogError("add next hop fail,the err is:the fibEntry is not found")
		return MakeControlResponse(411, "the fibEntry is not found", "")
	}
	// 删除这个标识前缀对应 FIB表项中的某个下一跳
	fibEntry.RemoveNextHop(face)
	if !fibEntry.HasNextHops() {
		// 如果空 直接删除整个表项
		err := f.fib.EraseByFIBEntry(fibEntry)
		if err != nil {
			common.LogError("add next hop fail,the err is", err)
			return MakeControlResponse(412, err.Error(), "")
		}
	}
	// 返回成功
	common.LogInfo("remove next hop success")
	return MakeControlResponse(200, "remove next hop success", "")
}

// ListEntries
// 获取fib表中所有的下一跳信息
//
// @Description:获取fib表中所有信息，并分片发送给客户端
// @receiver f
//
func (f *FibManager) ListEntries(topPrefix *component.Identifier, interest *packet.Interest, context *StatusDatasetContext) {
	//var response *mgmt.ControlResponse
	//fibEntrys := f.fib.GetAllEntry()
	//data, err := json.Marshal(fibEntrys)
	//if err != nil {
	//	common.LogError("get fib info fail,the err is", err)
	//	response = MakeControlResponse(400, "mashal fibEntrys fail , the err is:"+err.Error(), "")
	//	context.responseSender(response, interest)
	//	return
	//}
	//context.data = data
	//// 返回分片列表，并将分片放入缓存中去
	//dataList := context.Append()
	//if dataList == nil {
	//	common.LogError("get fib info fail,the err is", err)
	//	response = MakeControlResponse(400, "slice data packet err!", "")
	//	context.responseSender(response, interest)
	//	return
	//} else {
	//	common.LogInfo("get fib info success")
	//	for i, data := range dataList {
	//		// 包编码放在dataSender中
	//		context.dataSaver(data)
	//		if i == 0 {
	//			// 第一个包是包头 发送 其他包暂时存放在缓存 不发送 等待前端继续请求
	//			context.dataSender(data)
	//		}
	//	}
	//}
}
