//
// @Author: Jianming Que
// @Description:
// @Version: 1.0.0
// @Date: 2021/2/21 4:19 下午
// @Copyright: MIN-Group；国家重大科技基础设施——未来网络北大实验室；深圳市信息论与未来网络重点实验室
//
package plugin

//
// 全局的插件管理程序，主要包含以下功能
//
// @Description:
//	1. 注册插件；
//	2. 在 Forwarder 处理流程的各个锚点调用对应的回调，在调用时按顺序调用插件列表中对应的回调函数
//
type GlobalPluginManager struct {
}
