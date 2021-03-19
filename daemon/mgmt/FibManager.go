//
// @Author: Jianming Que
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
	"mir-go/daemon/lf"
	"mir-go/daemon/table"
	"strconv"
)

type FibManager struct {
	fib *table.FIB
	logicfaceTable *lf.LogicFaceTable
}

// 顶级域前缀 授权验证	验证命令参数	回调函数
// 添加下一跳命令
func (f *FibManager) AddNextHop(topPrefix *component.Identifier, interest *packet.Interest,
	parameters *mgmt.ControlParameters)*mgmt.ControlResponse{
	prefix := parameters.GetName()
	logicfaceId := parameters.GetLogicfaceId()
	cost := parameters.GetCost()
	// 如果标识组件个数大于规定树的最大深度
	if prefix.Size()>f.fib.GetMaxDepth(){
		// 返回信息错误 为什么要返回错误信息
		return &mgmt.ControlResponse{414,"the prefix is too long ,cannot exceed "+strconv.Itoa(f.fib.GetMaxDepth())+"components"}
	}
	// 根据Id从table中取出 logicface
	face:=f.logicfaceTable.Get(logicfaceId)
	if face == nil{
		return &mgmt.ControlResponse{410,"the face is not found"}
	}
	// 执行添加下一跳命令 放入表中
	f.fib.AddOrUpdate(prefix,face,cost)
	return &mgmt.ControlResponse{200,"add next hop success"}
}

func (f *FibManager) RemoveNextHop(topPrefix *component.Identifier, interest *packet.Interest,
	parameters *mgmt.ControlParameters)*mgmt.ControlResponse{
	// 获取前缀
	prefix := parameters.GetName()
	logicfaceId := parameters.GetLogicfaceId()
	// 根据Id从table中取出 logicface
	face:=f.logicfaceTable.Get(logicfaceId)
	if face == nil{
		return &mgmt.ControlResponse{410,"the face is not found"}
	}
	fibEntry:=f.fib.FindExactMatch(prefix)
	if fibEntry == nil{
		return &mgmt.ControlResponse{411,"the fibEntry is not found"}
	}
	fibEntry.RemoveNextHop(face)
	if !fibEntry.HasNextHops(){
		// 如果空 直接删除整个表项
		f.fib.EraseByFIBEntry(fibEntry)
	}
	// 返回成功
	return &mgmt.ControlResponse{200,"remove next hop success"}
}

func (f *FibManager) ListEntries(topPrefix *component.Identifier,interest *packet.Interest,context *StatusDatasetContext){


}

func (f *FibManager )sendControlResponse(response *mgmt.ControlResponse, interest *packet.Interest){


}