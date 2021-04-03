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
	"minlib/component"
	"minlib/mgmt"
	"minlib/packet"
	"mir-go/daemon/common"
	"mir-go/daemon/lf"
	"net"
	"strconv"
)

const (
	EtherLogicFace = iota
	TcpLogicFace
	UdpLogicFace
	UnixLogicFace
)

//
// face管理模块结构体
//
// @Description:face管理模块结构体
//
type FaceManager struct {
}

//
// 创建face管理模块函数
//
// @Description:创建face管理模块函数并返回指针
//
func CreateFaceManager() *FaceManager {
	return &FaceManager{}
}

//
// face管理模块初始化注册命令
//
// @Description:face管理模块初始化注册命令，包括create、destroy、list
// @receiver f
//
func (f *FaceManager) Init() {
	identifier, _ := component.CreateIdentifierByString("/min-mir/mgmt/localhost/face-mgmt/create")
	err := dispatcher.AddControlCommand(identifier, authorization, f.ValidateParameters, f.createFace)
	if err != nil {
		common.LogError("face add create-command fail,the err is:", err)
	}

	identifier, _ = component.CreateIdentifierByString("/min-mir/mgmt/localhost/face-mgmt/destroy")
	err = dispatcher.AddControlCommand(identifier, authorization, f.ValidateParameters, f.destroyFace)
	if err != nil {
		common.LogError("face add destroy-command fail,the err is:", err)
	}

	identifier, _ = component.CreateIdentifierByString("/min-mir/mgmt/localhost/face-mgmt/list")
	err = dispatcher.AddStatusDataset(identifier, authorization, f.listFaces)
	if err != nil {
		common.LogError("face add list-command fail,the err is:", err)
	}
}

//
// 创建连接face函数
//
// @Description:创建连接face函数，有Ether、TCP、UDP、UNIX四种
// @receiver f
// @Return:*mgmt.ControlResponse返回创建结果
//
func (f *FaceManager) createFace(topPrefix *component.Identifier, interest *packet.Interest,
	parameters *mgmt.ControlParameters) *mgmt.ControlResponse {
	uriScheme := parameters.ControlParameterUriScheme.UriScheme()
	uri := parameters.ControlParameterUri.Uri()
	localUri := parameters.ControlParameterLocalUri.LocalUri()
	switch uriScheme {
	case EtherLogicFace:
		remoteMacAddr, err := net.ParseMAC(uri)
		if err != nil {
			return &mgmt.ControlResponse{Code: 400, Msg: "parse remote address fail,the err is:" + err.Error()}
		}
		logicFaceId, err := lf.CreateEtherLogicFace(localUri, remoteMacAddr)
		if err != nil {
			return &mgmt.ControlResponse{Code: 400, Msg: "create EtherLogicFace fail,the err is:" + err.Error()}
		}
		return &mgmt.ControlResponse{Code: 200, Msg: "create face success,the id is " + strconv.FormatUint(logicFaceId, 10)}
	case TcpLogicFace:
		logicFaceId, err := lf.CreateTcpLogicFace(uri)
		if err != nil {
			return &mgmt.ControlResponse{Code: 400, Msg: "create TcpLogicFace fail,the err is:" + err.Error()}
		}
		return &mgmt.ControlResponse{Code: 200, Msg: "create face success,the id is " + strconv.FormatUint(logicFaceId, 10)}
	case UdpLogicFace:
		logicFaceId, err := lf.CreateUdpLogicFace(uri)
		if err != nil {
			return &mgmt.ControlResponse{Code: 400, Msg: "create TcpLogicFace fail,the err is:" + err.Error()}
		}
		return &mgmt.ControlResponse{Code: 200, Msg: "create face success,the id is " + strconv.FormatUint(logicFaceId, 10)}
	case UnixLogicFace:
		logicFaceId, err := lf.CreateUnixLogicFace(uri)
		if err != nil {
			return &mgmt.ControlResponse{Code: 400, Msg: "create TcpLogicFace fail,the err is:" + err.Error()}
		}
		return &mgmt.ControlResponse{Code: 200, Msg: "create face success,the id is " + strconv.FormatUint(logicFaceId, 10)}
	default:
		return &mgmt.ControlResponse{Code: 400, Msg: "Unsupported protocol"}
	}
}

//
// 根据LogicfaceId从全局FaceTable中删除face
//
// @Description:根据LogicfaceId从全局FaceTable中删除face
// @receiver f
// @Return:*mgmt.ControlResponse返回删除结果
//
func (f *FaceManager) destroyFace(topPrefix *component.Identifier, interest *packet.Interest,
	parameters *mgmt.ControlParameters) *mgmt.ControlResponse {
	logicfaceId := parameters.ControlParameterLogicFaceId.LogicFaceId()
	face := lf.GLogicFaceTable.GetLogicFacePtrById(logicfaceId)
	if face == nil {
		return &mgmt.ControlResponse{Code: 400, Msg: "the face is not existed"}
	}
	lf.GLogicFaceTable.RemoveByLogicFaceId(logicfaceId)
	return &mgmt.ControlResponse{Code: 200, Msg: "ok"}
}

//
// 获取所有的逻辑face并分片发送给客户端
//
// @Description:获取所有的逻辑face并分片发送给客户端
// @receiver f
//
func (f *FaceManager) listFaces(topPrefix *component.Identifier, interest *packet.Interest,
	context *StatusDatasetContext) {
	faceList := lf.GLogicFaceTable.GetAllFaceList()
	data, err := json.Marshal(faceList)
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

//
// face管理模块的参数验证函数
//
// @Description:face管理模块的参数验证函数，条件语句中的为必需字段，若有一项不合规范则返回false
// @receiver f
// @Return:bool
//
func (f *FaceManager) ValidateParameters(parameters *mgmt.ControlParameters) bool {
	if parameters.ControlParameterUri.IsInitial() &&
		parameters.ControlParameterLocalUri.IsInitial() &&
		parameters.ControlParameterUriScheme.IsInitial() {
		return true
	}
	return false
}
