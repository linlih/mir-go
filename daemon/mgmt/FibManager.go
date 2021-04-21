// Package mgmt
// @Author: yzy
// @Description:
// @Version: 1.0.0
// @Date: 2021/3/10 3:13 下午
// @Copyright: MIN-Group；国家重大科技基础设施——未来网络北大实验室；深圳市信息论与未来网络重点实验室
//
package mgmt

import (
	"github.com/sirupsen/logrus"
	"minlib/common"
	"minlib/component"
	"minlib/mgmt"
	"minlib/packet"
	"mir-go/daemon/lf"
	"mir-go/daemon/table"
	"strconv"
)

type FibInfo struct {
	Identifier   string
	NextHopsInfo []NextHopInfo
}

type NextHopInfo struct {
	LogicFaceId uint64
	Cost        uint64
}

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
	// /fib-mgmt/add => 添加一个转发表项
	identifier, _ := component.CreateIdentifierByString("/" + mgmt.ManagementModuleFibMgmt + "/" + mgmt.FibManagementActionAdd)
	// 绑定控制函数，加入参数验证
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
	// /fib-mgmt/del => 删除一个转发表项
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
	// /fib-mgmt/list => 展示所有转发表条目
	identifier, _ = component.CreateIdentifierByString("/" + mgmt.ManagementModuleFibMgmt + "/" + mgmt.FibManagementActionList)
	err = dispatcher.AddStatusDataset(identifier, dispatcher.authorization, f.ListEntries)
	if err != nil {
		common.LogError("add list-command fail,the err is:", err)
	}

	// /fib-mgmt/register => 注册一个前缀监听
	identifier, _ = component.CreateIdentifierByString("/" + mgmt.ManagementModuleFibMgmt + "/" + mgmt.FibManagementActionRegister)
	err = dispatcher.AddControlCommand(identifier, dispatcher.authorization, func(parameters *component.ControlParameters) bool {
		return parameters.ControlParameterPrefix.IsInitial()
	}, f.RegisterPrefix)
	if err != nil {
		common.LogError("add register-command fail,the err is:", err)
	}
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
		common.LogDebugWithFields(logrus.Fields{
			"max depth":          table.MAX_DEPTH,
			"the size of prefix": prefix.Size(),
		}, "the prefix is too long")
		// 返回前缀太长的错误信息
		return MakeControlResponse(400, "the prefix is too long ,cannot exceed "+strconv.Itoa(table.MAX_DEPTH)+"components", "")
	}

	// 根据Id从table中取出 LogicFace
	face := f.logicFaceTable.GetLogicFacePtrById(logicFaceId)
	if face == nil {
		common.LogDebugWithFields(logrus.Fields{
			"prefix":      prefix.ToUri(),
			"logicFaceId": strconv.FormatUint(logicFaceId, 10),
		}, "the logicFace is not existed")
		return MakeControlResponse(400, "the face is not found", "")
	}
	// 查找前缀是否存在 如果存在而且是只读的话 那么不可以改变
	fibEntry := f.fib.FindExactMatch(prefix)
	if fibEntry != nil && !fibEntry.IsChanged() {
		common.LogDebugWithFields(logrus.Fields{
			"prefix": fibEntry.GetIdentifier().ToUri(),
		}, "change read only prefix")
		return MakeControlResponse(400, "read only,the prefix can't be changed", "")
	}
	f.fib.AddOrUpdate(prefix, face, cost)
	common.LogInfo("add next hop success")
	return MakeControlResponse(200, "add next hop success", "")
}

// RemoveNextHop
// 根据logicFace在fib表中删除下一跳
//
// @Description:根据logicFace在fib表中删除下一跳并返回删除结果
// @receiver f
//
func (f *FibManager) RemoveNextHop(topPrefix *component.Identifier, interest *packet.Interest,
	parameters *component.ControlParameters) *mgmt.ControlResponse {
	// 获取前缀
	prefix := parameters.ControlParameterPrefix.Prefix()
	logicFaceId := parameters.ControlParameterLogicFaceId.LogicFaceId()

	// 根据Id从table中取出 logicFace
	face := f.logicFaceTable.GetLogicFacePtrById(logicFaceId)
	if face == nil {
		common.LogDebugWithFields(logrus.Fields{
			"logicFaceId": strconv.FormatUint(logicFaceId, 10),
		}, "the logicFace is not existed")
		return MakeControlResponse(400, "the face is not found", "")

	}
	fibEntry := f.fib.FindExactMatch(prefix)
	if fibEntry == nil {
		common.LogDebugWithFields(logrus.Fields{
			"prefix": prefix.ToUri(),
		}, "fibEntry is not found")
		return MakeControlResponse(400, "the fibEntry is not found", "")
	}
	// 删除这个标识前缀对应 FIB表项中的某个下一跳
	if !fibEntry.IsChanged() {
		common.LogDebugWithFields(logrus.Fields{
			"prefix": fibEntry.GetIdentifier().ToUri(),
		}, "change read only prefix")
		return MakeControlResponse(400, "read only,the prefix can't be changed", "")
	}
	fibEntry.RemoveNextHop(face)
	if !fibEntry.HasNextHops() {
		// 如果空 直接删除整个表项
		err := f.fib.EraseByFIBEntry(fibEntry)
		if err != nil {
			common.LogDebugWithFields(logrus.Fields{
				"error": err,
			}, "delete the fibEntry fail")
			return MakeControlResponse(400, err.Error(), "")
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
	fibEntryList := f.fib.GetAllEntry()
	for _, fibEntry := range fibEntryList {
		var nextHopInfo []NextHopInfo
		for _, nextHop := range fibEntry.GetNextHops() {
			nextHopInfo = append(nextHopInfo, NextHopInfo{LogicFaceId: nextHop.LogicFace.LogicFaceId, Cost: nextHop.Cost})
		}
		fibInfo := FibInfo{
			Identifier:   fibEntry.GetIdentifier().ToString(),
			NextHopsInfo: nextHopInfo,
		}
		context.Append(fibInfo)
	}
	// 获取当前表的版本号
	currentVersion := f.fib.GetVersion()

	_ = context.Done(currentVersion)
}

// NextHopCleaner
// 从fib表项中清除所有以指定id为下一跳的nextHop
//
// @Description:从fib表项中清除所有以指定id为下一跳的nextHop
// @Parameters: logicFaceId uint64
// @receiver f
//
func (f *FibManager) NextHopCleaner(logicFaceId uint64) {
	fibEntryList := f.fib.GetAllEntry()
	for _, fibEntry := range fibEntryList {
		fibEntry.RWlock.Lock()
		delete(fibEntry.NextHopList, logicFaceId)
		if len(fibEntry.NextHopList) == 0 {
			f.fib.EraseByFIBEntry(fibEntry)
		}
		fibEntry.RWlock.Unlock()
	}
	common.LogInfo("the face is deleted,clean next hop ---------------------------- ")
}

// RegisterPrefix 处理注册前缀
//
// @Description:
// @receiver f
// @param topPrefix
// @param interest
// @param parameters
// @return *mgmt.ControlResponse
//
func (f *FibManager) RegisterPrefix(topPrefix *component.Identifier, interest *packet.Interest,
	parameters *component.ControlParameters) *mgmt.ControlResponse {
	parameters.SetLogicFaceId(interest.IncomingLogicFaceId.GetIncomingLogicFaceId())
	return f.AddNextHop(topPrefix, interest, parameters)
}
