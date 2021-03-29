//
// @Author: yzy
// @Description:
// @Version: 1.0.0
// @Date: 2021/3/10 3:13 下午
// @Copyright: MIN-Group；国家重大科技基础设施——未来网络北大实验室；深圳市信息论与未来网络重点实验室
//
package mgmt

import (
	"encoding/json"
	"fmt"
	"minlib/component"
	"minlib/mgmt"
	"minlib/packet"
	"mir-go/daemon/lf"
	"mir-go/daemon/table"
	"strconv"
)

type FibManager struct {
	fib *table.FIB
}

func CreateFibManager() *FibManager {
	return &FibManager{
		fib: table.CreateFIB(),
	}
}

// 注册命令 一个前缀对应一个命令
func (f *FibManager) Init() {
	identifier, _ := component.CreateIdentifierByString("/fib-mgmt/add")
	err := dispatcher.AddControlCommand(identifier, authorization, f.ValidateParameters, f.AddNextHop)
	if err != nil {
		fmt.Println("add add-command fail,the err is:", err)
	}
	identifier, _ = component.CreateIdentifierByString("/fib-mgmt/delete")
	err = dispatcher.AddControlCommand(identifier, authorization, f.ValidateParameters, f.RemoveNextHop)
	if err != nil {
		fmt.Println("add delete-command fail,the err is:", err)
	}
	identifier, _ = component.CreateIdentifierByString("/fib-mgmt/list")
	err = dispatcher.AddStatusDataset(identifier, authorization, f.ListEntries)
	if err != nil {
		fmt.Println("add list-command fail,the err is:", err)
	}
}

// 在 FIB 表中添加一个到指定前缀的路由
// mirc fib add identifier <IDENTIFIER> nexthop <LFID> [cost <COST>]

func (f *FibManager) AddNextHop(topPrefix *component.Identifier, interest *packet.Interest,
	parameters *mgmt.ControlParameters) *mgmt.ControlResponse {
	prefix := parameters.Prefix
	logicfaceId := parameters.LogicfaceId
	cost := parameters.Cost
	// 标识前缀 不能太长 太长返回错误信息
	if prefix.Size() > table.MAX_DEPTH {
		// 返回前缀太长的错误信息
		return &mgmt.ControlResponse{Code: 414, Msg: "the prefix is too long ,cannot exceed " + strconv.Itoa(table.MAX_DEPTH) + "components"}
	}
	// 根据Id从table中取出 logicface
	face := lf.GLogicFaceTable.GetLogicFacePtrById(logicfaceId)
	if face == nil {
		fmt.Println(prefix.ToUri() + " logicfaceId:" + strconv.FormatUint(logicfaceId, 10) + "failed!")
		return &mgmt.ControlResponse{Code: 410, Msg: "the face is not found"}
	}
	// 执行添加下一跳命令 放入表中
	f.fib.AddOrUpdate(prefix, face, cost)
	return &mgmt.ControlResponse{Code: 200, Msg: "add next hop success"}
}

func (f *FibManager) RemoveNextHop(topPrefix *component.Identifier, interest *packet.Interest,
	parameters *mgmt.ControlParameters) *mgmt.ControlResponse {
	// 获取前缀
	prefix := parameters.Prefix
	logicfaceId := parameters.LogicfaceId
	// 根据Id从table中取出 logicface
	face := lf.GLogicFaceTable.GetLogicFacePtrById(logicfaceId)
	if face == nil {
		return &mgmt.ControlResponse{Code: 410, Msg: "the face is not found"}
	}
	fibEntry := f.fib.FindExactMatch(prefix)
	if fibEntry == nil {
		return &mgmt.ControlResponse{Code: 411, Msg: "the fibEntry is not found"}
	}
	// 删除这个标识前缀对应 FIB表项中的某个下一跳
	fibEntry.RemoveNextHop(face)
	if !fibEntry.HasNextHops() {
		// 如果空 直接删除整个表项
		err := f.fib.EraseByFIBEntry(fibEntry)
		if err != nil {
			fmt.Println(err)
			return &mgmt.ControlResponse{Code: 412, Msg: err.Error()}
		}
	}
	// 返回成功
	return &mgmt.ControlResponse{Code: 200, Msg: "remove next hop success"}
}

func (f *FibManager) ListEntries(topPrefix *component.Identifier, interest *packet.Interest,
	context *StatusDatasetContext) {
	fibEntrys := f.fib.GetAllEntry()
	data, err := json.Marshal(fibEntrys)
	if err != nil {
		res := &mgmt.ControlResponse{Code: 400, Msg: "mashal fibEntrys fail , the err is:" + err.Error()}
		context.nackSender(res, interest)
		return
	}
	res := &mgmt.ControlResponse{Code: 200, Msg: "", Data: string(data)}
	newData, err := json.Marshal(res)
	if err != nil {
		res = &mgmt.ControlResponse{Code: 400, Msg: "mashal fibEntrys fail , the err is:" + err.Error()}
		context.nackSender(res, interest)
		return
	}
	context.data = newData
}

// FibManager验证参数函数 返回是否验证成功
func (f *FibManager) ValidateParameters(parameters *mgmt.ControlParameters) bool {
	if parameters.Prefix != nil && parameters.Cost > 0 && parameters.LogicfaceId > 0 {
		return true
	}
	return false
}
