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
	"mir-go/daemon/table"
)

type CsManager struct {
	cs          *table.CS
	enableServe bool //	是否可以展示信息
	enableAdd   bool // 是否可以添加缓存
}

const ERASE_LIMIT = 256

func CreateCsManager() *CsManager {
	return &CsManager{
		cs:          new(table.CS),
		enableServe: false,
		enableAdd:   false,
	}
}

func (c *CsManager) Init() {
	identifier, _ := component.CreateIdentifierByString("/min-mir/mgmt/localhost/cs-mgmt/delete")
	err := dispatcher.AddControlCommand(identifier, authorization, c.ValidateParameters, c.changeConfig)
	if err != nil {
		fmt.Println("cs add delete-command fail,the err is:", err)
	}
	identifier, _ = component.CreateIdentifierByString("/min-mir/mgmt/localhost/cs-mgmt/list")
	err = dispatcher.AddStatusDataset(identifier, authorization, c.serveInfo)
	if err != nil {
		fmt.Println("cs add list-command fail,the err is:", err)
	}
}

func (c *CsManager) changeConfig(topPrefix *component.Identifier, interest *packet.Interest,
	parameters *mgmt.ControlParameters) *mgmt.ControlResponse {
	c.enableServe = true
	return nil
}

func (c *CsManager) serveInfo(topPrefix *component.Identifier, interest *packet.Interest,
	context *StatusDatasetContext) {
	if c.enableServe {
		var CSInfo = struct {
			enableServe bool
			enableAdd   bool
			size        uint64
			hits        uint64
			misses      uint64
		}{
			enableServe: c.enableServe,
			enableAdd:   c.enableAdd,
			size:        c.cs.Size(),
			hits:        c.cs.Hits,
			misses:      c.cs.Misses,
		}
		data, err := json.Marshal(CSInfo)
		if err != nil {
			res := &mgmt.ControlResponse{Code: 400, Msg: "mashal CSInfo fail , the err is:" + err.Error()}
			context.nackSender(res, interest)
		}
		res := &mgmt.ControlResponse{Code: 200, Msg: "", Data: string(data)}
		newData, err := json.Marshal(res)
		if err != nil {
			res = &mgmt.ControlResponse{Code: 400, Msg: "mashal CSInfo fail , the err is:" + err.Error()}
			context.nackSender(res, interest)
			return
		}
		context.data = newData
	}
	res := &mgmt.ControlResponse{Code: 400, Msg: "have no Permission to get CsInfo!"}
	context.nackSender(res, interest)
}

func (c *CsManager) ValidateParameters(parameters *mgmt.ControlParameters) bool {
	if parameters.Prefix != nil && parameters.Count > 0 && parameters.Capacity > 0 {
		return true
	}
	return false
}
