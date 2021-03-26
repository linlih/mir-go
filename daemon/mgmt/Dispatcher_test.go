package mgmt

import (
	"encoding/json"
	"fmt"
	"minlib/component"
	"minlib/mgmt"
	"minlib/packet"
	"testing"
)

func Test(t *testing.T) {
	interest := &packet.Interest{}
	Prefix, _ := component.CreateIdentifierByString("/fib-mgmt/add")
	para := &mgmt.ControlParameters{LogicfaceId: 1, Cost: 1, Prefix: Prefix}
	byteArrary, _ := json.Marshal(para)
	topPrefix, _ := component.CreateIdentifierByString("/min-mir/mgmt/localhost")
	identifier := *topPrefix
	components := component.CreateIdentifierComponentByByteArray(byteArrary)
	identifier.Append(components)
	interest.SetName(&identifier)
	module := dispatcher.module[Prefix.ToUri()]
	if module.authorization(topPrefix, interest, para, authorizationAccept, authorizationReject) {
		//普通查询
		res := module.ccHandler(topPrefix, interest, para)
		fmt.Println(res)
		dispatcher.sendControlResponse(res, interest)
		////表项查询 数据量大需要分片 自定义没有找到进行的操作
		//dispatcher.queryStorage(topPrefix, interest, func(topPrefix *component.Identifier, interest *packet.Interest) {
		//	var context = CreateSDC(interest, dispatcher.sendData, dispatcher.sendControlResponse)
		//	module.sdHandler(topPrefix, interest, context)
		//	// 放入缓存
		//	context.Append()
		//	// 发送Data数据包
		//})
	}
}
