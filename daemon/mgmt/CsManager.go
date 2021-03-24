//
// @Author: Jianming Que
// @Description:
// @Version: 1.0.0
// @Date: 2021/3/10 3:13 下午
// @Copyright: MIN-Group；国家重大科技基础设施——未来网络北大实验室；深圳市信息论与未来网络重点实验室
//
package mgmt

import (
	"fmt"
	"minlib/component"
	"minlib/mgmt"
	"minlib/packet"
	"mir-go/daemon/table"
)

type CsManager struct {
	cs *table.CS
}

const ERASE_LIMIT = 256

// 注册命令 一个前缀对应一个命令
func (c *CsManager) Init() {
	identifier, _ := component.CreateIdentifierByString("/min-mir/mgmt/cs-mgmt/localhost/change")
	err := dispatcher.AddControlCommand(identifier, authorization, c.ValidateParameters, c.changeConfig)
	if err != nil {
		fmt.Println("cs add change-command fail,the err is:", err)
	}
	identifier, _ = component.CreateIdentifierByString("/min-mir/mgmt/localhost/cs-mgmt/delete")
	err = dispatcher.AddControlCommand(identifier, authorization, c.ValidateParameters, c.erase)
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
	parameters *mgmt.ControlParameters) *mgmt.ControlResponse{

}

func (c *CsManager) erase(topPrefix *component.Identifier, interest *packet.Interest,
	parameters *mgmt.ControlParameters) *mgmt.ControlResponse{

}

func (c *CsManager) serveInfo(topPrefix *component.Identifier, interest *packet.Interest,
	context *StatusDatasetContext){

}

func (c *CsManager) ValidateParameters(parameters *mgmt.ControlParameters) bool {
	if parameters.Prefix != nil && parameters.Cost > 0 && parameters.LogicfaceId > 0 {
		return true
	}
	return false
}