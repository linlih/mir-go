//
// @Author: Jianming Que
// @Description:
// @Version: 1.0.0
// @Date: 2021/3/15 8:58 下午
// @Copyright: MIN-Group；国家重大科技基础设施——未来网络北大实验室；深圳市信息论与未来网络重点实验室
//
package fw

import (
	"minlib/packet"
	"mir-go/daemon/lf"
	"mir-go/daemon/table"
)

//
// 最佳路由转发策略实现
//
// @Description:
//
type BestRouteStrategy struct {
	StrategyBase
}

func (brs *BestRouteStrategy) AfterReceiveInterest(ingress *lf.LogicFace, interest *packet.Interest, pitEntry *table.PITEntry) {

}

func (brs *BestRouteStrategy) AfterReceiveNack(ingress *lf.LogicFace, nack *packet.Nack, pitEntry *table.PITEntry) {

}

func (brs *BestRouteStrategy) AfterReceiveCPacket(ingress *lf.LogicFace, cPacket *packet.CPacket) {

}
