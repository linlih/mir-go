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
	"net"
	"strconv"
)

const (
	EtherLogicFace = iota
	TcpLogicFace
	UdpLogicFace
	UnixLogicFace
)

type FaceManager struct {
}

func CreateFaceManager() *FaceManager {
	return &FaceManager{}
}

func (f *FaceManager) Init() {
	identifier, _ := component.CreateIdentifierByString("/min-mir/mgmt/localhost/face-mgmt/create")
	err := dispatcher.AddControlCommand(identifier, authorization, f.ValidateParameters, f.createFace)
	if err != nil {
		fmt.Println("face add create-command fail,the err is:", err)
	}

	identifier, _ = component.CreateIdentifierByString("/min-mir/mgmt/localhost/face-mgmt/destroy")
	err = dispatcher.AddControlCommand(identifier, authorization, f.ValidateParameters, f.destroyFace)
	if err != nil {
		fmt.Println("face add destroy-command fail,the err is:", err)
	}

	identifier, _ = component.CreateIdentifierByString("/min-mir/mgmt/localhost/face-mgmt/list")
	err = dispatcher.AddStatusDataset(identifier, authorization, f.listFaces)
	if err != nil {
		fmt.Println("face add list-command fail,the err is:", err)
	}
}

//
func (f *FaceManager) createFace(topPrefix *component.Identifier, interest *packet.Interest,
	parameters *mgmt.ControlParameters) *mgmt.ControlResponse {

	switch parameters.Scheme {
	case EtherLogicFace:
		remoteMacAddr, err := net.ParseMAC(parameters.RemoteUri)
		if err != nil {
			return &mgmt.ControlResponse{Code: 400, Msg: "parse remote address fail,the err is:" + err.Error()}
		}
		logicFaceId, err := lf.CreateEtherLogicFace(parameters.LocalUri, remoteMacAddr)
		if err != nil {
			return &mgmt.ControlResponse{Code: 400, Msg: "create EtherLogicFace fail,the err is:" + err.Error()}
		}
		return &mgmt.ControlResponse{Code: 200, Msg: "create face success,the id is " + strconv.FormatUint(logicFaceId, 10)}
	case TcpLogicFace:
		logicFaceId, err := lf.CreateTcpLogicFace(parameters.RemoteUri)
		if err != nil {
			return &mgmt.ControlResponse{Code: 400, Msg: "create TcpLogicFace fail,the err is:" + err.Error()}
		}
		return &mgmt.ControlResponse{Code: 200, Msg: "create face success,the id is " + strconv.FormatUint(logicFaceId, 10)}
	case UdpLogicFace:
		logicFaceId, err := lf.CreateUdpLogicFace(parameters.RemoteUri)
		if err != nil {
			return &mgmt.ControlResponse{Code: 400, Msg: "create TcpLogicFace fail,the err is:" + err.Error()}
		}
		return &mgmt.ControlResponse{Code: 200, Msg: "create face success,the id is " + strconv.FormatUint(logicFaceId, 10)}
	case UnixLogicFace:
		logicFaceId, err := lf.CreateUnixLogicFace(parameters.RemoteUri)
		if err != nil {
			return &mgmt.ControlResponse{Code: 400, Msg: "create TcpLogicFace fail,the err is:" + err.Error()}
		}
		return &mgmt.ControlResponse{Code: 200, Msg: "create face success,the id is " + strconv.FormatUint(logicFaceId, 10)}
	default:
		return &mgmt.ControlResponse{Code: 400, Msg: "Unsupported protocol"}
	}
}

func (f *FaceManager) destroyFace(topPrefix *component.Identifier, interest *packet.Interest,
	parameters *mgmt.ControlParameters) *mgmt.ControlResponse {
	face := lf.GLogicFaceTable.GetLogicFacePtrById(parameters.LogicfaceId)
	if face == nil {
		return &mgmt.ControlResponse{Code: 400, Msg: "the face is not existed"}
	}
	lf.GLogicFaceTable.RemoveByLogicFaceId(parameters.LogicfaceId)
	return &mgmt.ControlResponse{Code: 200, Msg: "ok"}
}

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

func (f *FaceManager) ValidateParameters(parameters *mgmt.ControlParameters) bool {
	if parameters.RemoteUri != "" && parameters.LocalUri != "" && parameters.Scheme != 0 {
		return true
	}
	return false
}
