// Package mgmt
// @Author: yzy
// @Description:
// @Version: 1.0.0
// @Date: 2021/3/10 3:13 下午
// @Copyright: MIN-Group；国家重大科技基础设施——未来网络北大实验室；深圳市信息论与未来网络重点实验室
//
package mgmt

import (
	"encoding/json"
	"minlib/common"
	"minlib/component"
	"minlib/mgmt"
	"minlib/packet"
	"mir-go/daemon/lf"
	"net"
	"strconv"
)

type FaceInfo struct {
	LogicFaceId uint64
	RemoteUri   string
	LocalUri    string
	Mtu         int
}

// FaceManager face管理模块结构体
//
// @Description:face管理模块结构体
//
type FaceManager struct {
	logicFaceTable *lf.LogicFaceTable
}

// CreateFaceManager
// 创建face管理模块函数
//
// @Description:创建face管理模块函数并返回指针
//
func CreateFaceManager() *FaceManager {
	return &FaceManager{}
}

// Init
// face管理模块初始化注册命令
//
// @Description:face管理模块初始化注册命令，包括create、destroy、list
// @receiver f
//
func (f *FaceManager) Init(dispatcher *Dispatcher, logicFaceTable *lf.LogicFaceTable) {
	f.logicFaceTable = logicFaceTable
	identifier, _ := component.CreateIdentifierByString("/face-mgmt/add")
	err := dispatcher.AddControlCommand(identifier, dispatcher.authorization, func(parameters *mgmt.ControlParameters) bool {
		if parameters.ControlParameterUri.IsInitial() &&
			parameters.ControlParameterLocalUri.IsInitial() &&
			parameters.ControlParameterMtu.IsInitial() &&
			parameters.ControlParameterLogicFacePersistency.IsInitial() {
			return true
		}
		return false
	}, f.createFace)
	if err != nil {
		common.LogError("face add create-command fail,the err is:", err)
	}

	identifier, _ = component.CreateIdentifierByString("/face-mgmt/destroy")
	err = dispatcher.AddControlCommand(identifier, dispatcher.authorization, func(parameters *mgmt.ControlParameters) bool {
		if parameters.ControlParameterLogicFaceId.IsInitial() {
			return true
		}
		return false
	}, f.destroyFace)
	if err != nil {
		common.LogError("face add destroy-command fail,the err is:", err)
	}

	identifier, _ = component.CreateIdentifierByString("/face-mgmt/list")
	err = dispatcher.AddStatusDataset(identifier, dispatcher.authorization, f.listFaces)
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
	case component.ControlParameterUriSchemeEther:
		remoteMacAddr, err := net.ParseMAC(uri)
		if err != nil {
			common.LogError("create face fail!the err is:", err)
			return MakeControlResponse(400, "parse remote address fail,the err is:"+err.Error(), "")

		}
		logicFaceId, err := lf.CreateEtherLogicFace(localUri, remoteMacAddr)
		if err != nil {
			common.LogError("create face fail!the err is:", err)
			return MakeControlResponse(400, "create EtherLogicFace fail,the err is:"+err.Error(), "")
		} else {
			common.LogInfo("create face success")
			return MakeControlResponse(200, "create face success,the id is "+strconv.FormatUint(logicFaceId, 10), "")
		}
	case component.ControlParameterUriSchemeTCP:
		logicFaceId, err := lf.CreateTcpLogicFace(uri)
		if err != nil {
			common.LogError("create face fail!the err is:", err)
			return MakeControlResponse(400, "create TcpLogicFace fail,the err is:"+err.Error(), "")
		} else {
			common.LogInfo("create face success")
			return MakeControlResponse(200, "create face success,the id is "+strconv.FormatUint(logicFaceId, 10), "")
		}
	case component.ControlParameterUriSchemeUDP:
		logicFaceId, err := lf.CreateUdpLogicFace(uri)
		if err != nil {
			common.LogError("create face fail!the err is:", err)
			return MakeControlResponse(400, "create UdpLogicFace fail,the err is:"+err.Error(), "")
		}
		common.LogInfo("create face success")
		return MakeControlResponse(200, "create face success,the id is "+strconv.FormatUint(logicFaceId, 10), "")
	case component.ControlParameterUriSchemeUnix:
		logicFaceId, err := lf.CreateUnixLogicFace(uri)
		if err != nil {
			common.LogError("create face fail!the err is:", err)
			return MakeControlResponse(400, "create UnixLogicFace fail,the err is:"+err.Error(), "")
		}
		common.LogInfo("create face success")
		return MakeControlResponse(200, "create face success,the id is "+strconv.FormatUint(logicFaceId, 10), "")

	default:
		common.LogError("create face fail!the err is:Unsupported protocol")
		return MakeControlResponse(400, "Unsupported protocol", "")
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
	logicFaceId := parameters.ControlParameterLogicFaceId.LogicFaceId()
	face := f.logicFaceTable.GetLogicFacePtrById(logicFaceId)
	if face == nil {
		common.LogError("destory face fail,the err is: inner face is null")
		return MakeControlResponse(400, "the face is not existed", "")
	}
	f.logicFaceTable.RemoveByLogicFaceId(logicFaceId)
	common.LogInfo("destory face success")
	return MakeControlResponse(200, "destory face success!", "")
}

//
// 获取所有的逻辑face并分片发送给客户端
//
// @Description:获取所有的逻辑face并分片发送给客户端
// @receiver f
//
func (f *FaceManager) listFaces(topPrefix *component.Identifier, interest *packet.Interest, context *StatusDatasetContext) {
	var response *mgmt.ControlResponse
	faceList := f.logicFaceTable.GetAllFaceList()
	var faceInfoList []*FaceInfo
	for _, face := range faceList {
		faceInfo := &FaceInfo{
			LogicFaceId: face.LogicFaceId,
			RemoteUri:   face.GetRemoteUri(),
			LocalUri:    face.GetLocalUri(),
		}
		faceInfoList = append(faceInfoList, faceInfo)
	}
	data, err := json.Marshal(faceInfoList)
	if err != nil {
		common.LogError("get face info fail,the err is:", err)
		response = MakeControlResponse(400, "mashal fibEntrys fail , the err is:"+err.Error(), "")
		context.nackSender(response, interest)
	}
	context.data = data
	// 返回分片列表，并将分片放入缓存中去
	dataList := context.Append()
	if dataList == nil {
		common.LogError("get face info fail,the err is:", err)
		response = MakeControlResponse(400, "slice data packet err!", "")
		context.nackSender(response, interest)
		return
	} else {
		common.LogInfo("get face info success")
		for i, data := range dataList {
			// 包编码放在dataSender中
			context.dataSaver(data)
			if i == 0 {
				// 第一个包是包头 发送 其他包暂时存放在缓存 不发送 等待前端继续请求
				context.dataSender(data)
			}
		}
	}
}

// ValidateParameters
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
