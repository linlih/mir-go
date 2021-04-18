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
	"strings"
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

	// /face-mgmt/add => 添加一个逻辑接口
	identifier, _ := component.CreateIdentifierByString("/face-mgmt/add")
	err := dispatcher.AddControlCommand(identifier, dispatcher.authorization,
		func(parameters *component.ControlParameters) bool {
			if parameters.ControlParameterUri.IsInitial() &&
				parameters.ControlParameterMtu.IsInitial() &&
				parameters.ControlParameterLogicFacePersistency.IsInitial() {
				return true
			}
			return false
		},
		f.addLogicFace)
	if err != nil {
		common.LogError("face add create-command fail,the err is:", err)
	}

	// /face-mgmt/del => 删除一个逻辑接口
	identifier, _ = component.CreateIdentifierByString("/face-mgmt/del")
	err = dispatcher.AddControlCommand(identifier, dispatcher.authorization, func(parameters *component.ControlParameters) bool {
		if parameters.ControlParameterLogicFaceId.IsInitial() {
			return true
		}
		return false
	}, f.delLogicFace)
	if err != nil {
		common.LogError("face add destroy-command fail,the err is:", err)
	}

	// /face-mgmt/list => 获取所有逻辑接口
	identifier, _ = component.CreateIdentifierByString("/face-mgmt/list")
	err = dispatcher.AddStatusDataset(identifier, dispatcher.authorization, f.listLogicFace)
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
func (f *FaceManager) addLogicFace(topPrefix *component.Identifier, interest *packet.Interest,
	parameters *component.ControlParameters) *mgmt.ControlResponse {

	// 提取参数
	uriScheme := parameters.ControlParameterUriScheme.UriScheme()
	uri := parameters.ControlParameterUri.Uri()
	localUri := parameters.ControlParameterLocalUri.LocalUri()
	mtu := parameters.ControlParameterMtu.Mtu()
	persistency := parameters.ControlParameterLogicFacePersistency.Persistency()

	// 判断Uri格式是否正确
	uriItems := strings.Split(uri, "://")
	if len(uriItems) != 2 {
		return MakeControlResponse(400, "Remote uri is wrong, expect one '://' item, "+uri, "")
	}

	// 根据不同的 Uri scheme，创建不同的逻辑接口
	switch uriScheme {
	case component.ControlParameterUriSchemeEther:
		remoteMacAddr, err := net.ParseMAC(uri)
		if err != nil {
			return MakeControlResponse(400, "parse remote address fail,the err is:"+err.Error(), "")
		}
		logicFace, err := lf.CreateEtherLogicFace(localUri, remoteMacAddr)
		if err != nil || logicFace == nil {
			msg := ""
			if err != nil {
				msg = err.Error()
			}
			return MakeControlResponse(400, "create EtherLogicFace fail,the err is:"+msg, "")
		}
		logicFace.Mtu = mtu
		logicFace.Persistency = persistency
		return MakeControlResponse(200, "create face success,the id is "+strconv.FormatUint(logicFace.LogicFaceId, 10), "")
	case component.ControlParameterUriSchemeTCP:
		logicFace, err := lf.CreateTcpLogicFace(uriItems[1])
		if err != nil || logicFace == nil {
			msg := ""
			if err != nil {
				msg = err.Error()
			}
			return MakeControlResponse(400, "create TcpLogicFace fail,the err is:"+msg, "")
		}
		return MakeControlResponse(200, "create face success,the id is "+strconv.FormatUint(logicFace.LogicFaceId, 10), "")
	case component.ControlParameterUriSchemeUDP:
		logicFace, err := lf.CreateUdpLogicFace(uri)
		if err != nil || logicFace == nil {
			msg := ""
			if err != nil {
				msg = err.Error()
			}
			return MakeControlResponse(400, "create UdpLogicFace fail,the err is:"+msg, "")
		}
		return MakeControlResponse(200, "create face success,the id is "+strconv.FormatUint(logicFace.LogicFaceId, 10), "")
	case component.ControlParameterUriSchemeUnix:
		logicFace, err := lf.CreateUnixLogicFace(uri)
		if err != nil || logicFace == nil {
			msg := ""
			if err != nil {
				msg = err.Error()
			}
			return MakeControlResponse(400, "create UnixLogicFace fail,the err is:"+msg, "")
		}
		logicFace.Mtu = mtu
		logicFace.Persistency = persistency
		return MakeControlResponse(200, "create face success,the id is "+strconv.FormatUint(logicFace.LogicFaceId, 10), "")

	default:
		return MakeControlResponse(400, "Unsupported protocol", "")
	}
}

//
// 根据LogicFaceId从全局FaceTable中删除face
//
// @Description: 根据LogicFaceId从全局FaceTable中删除face
// @receiver f
// @Return:*mgmt.ControlResponse返回删除结果
//
func (f *FaceManager) delLogicFace(topPrefix *component.Identifier, interest *packet.Interest,
	parameters *component.ControlParameters) *mgmt.ControlResponse {
	logicFaceId := parameters.ControlParameterLogicFaceId.LogicFaceId()
	face := f.logicFaceTable.GetLogicFacePtrById(logicFaceId)
	if face == nil {
		common.LogError("del face fail,the err is: inner face is null")
		return MakeControlResponse(400, "the face is not existed", "")
	}
	f.logicFaceTable.RemoveByLogicFaceId(logicFaceId)
	common.LogInfo("del face success")
	return MakeControlResponse(200, "del face success!", "")
}

//
// 获取所有的逻辑face并分片发送给客户端
//
// @Description:获取所有的逻辑face并分片发送给客户端
// @receiver f
//
func (f *FaceManager) listLogicFace(topPrefix *component.Identifier, interest *packet.Interest, context *StatusDatasetContext) {
	var response *mgmt.ControlResponse

	// 得到逻辑接口
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
				data.NoCache.SetNoCache(true) // 元数据不缓存
				context.dataSender(data)
			}
		}
	}
}
