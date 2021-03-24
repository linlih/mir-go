//
// @Author: Jianming Que
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
	"mir-go/daemon/lf"
)

type FaceManager struct {

}

func (f *FaceManager) createFace(topPrefix *component.Identifier, interest *packet.Interest,
	parameters *mgmt.ControlParameters) *mgmt.ControlResponse {



}



func (f *FaceManager) destroyFace(topPrefix *component.Identifier, interest *packet.Interest,
	parameters *mgmt.ControlParameters) *mgmt.ControlResponse{
	face:=lf.GLogicFaceTable.GetLogicFacePtrById(parameters.LogicfaceId)
	if face == nil{
		return &mgmt.ControlResponse{Code: 400,Msg: "the face is not existed"}
	}
	lf.GLogicFaceTable.RemoveByLogicFaceId(parameters.LogicfaceId)
	return &mgmt.ControlResponse{Code: 200,Msg: "ok"}
}

func (f *FaceManager) listFaces(topPrefix *component.Identifier, interest *packet.Interest,
	context *StatusDatasetContext){

	faceList:=[]*lf.LogicFace
	for _,v:=range lf.GLogicFaceTable{

	}
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

