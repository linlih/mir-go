//
// @Author: Jianming Que
// @Description:
// @Version: 1.0.0
// @Date: 2021/2/22 12:07 下午
// @Copyright: MIN-Group；国家重大科技基础设施——未来网络北大实验室；深圳市信息论与未来网络重点实验室
//
package fw

import "mir/daemon/lf"

type EndpointId uint64

type LogicFaceEndpoint struct {
	EndpointId
	logicFace *lf.LogicFace
}
