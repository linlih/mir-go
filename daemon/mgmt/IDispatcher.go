//
// @Author: yzy
// @Description:
// @Version: 1.0.0
// @Date: 2021/3/10 3:13 下午
// @Copyright: MIN-Group；国家重大科技基础设施——未来网络北大实验室；深圳市信息论与未来网络重点实验室
//
package mgmt

import (
	"minlib/component"
	"minlib/mgmt"
	"minlib/packet"
	"sync"
)

// 全局模块锁
var moduleLock sync.RWMutex

//
// 授权通过回调
//
// @Description:
//
type AuthorizationAccept func()

//
// 授权拒绝回调
//
// @Description:
//
type AuthorizationReject func(errorType int)

//
// 一个回调函数，用于对收到的控制命令进行授权验证
//
// @Description:
//
// @param topPrefix	顶级管理前缀，例如 "/min-mir-go/mgmt/localhost，可以通过本参数实现以下控制需求
//					1. 比如可以控制只有指定的顶级管理前缀可以授权通过，其它都不行，例如：/min-mir-go/mgmt/localhost 合法，其它前缀均不合法；
//					2. 也可以实现对不同的顶级前缀实现不同级别的授权，比如 "/min-mir-go/mgmt/localhost" 认为是本地管理员，拥有较高权限，默认
//					   可以控制和获取路由器状态；"/<路由器前缀>/min-mir-go/mgmt/remote" 认定为远程管理员发过来的命令，拥有较局限的权限，只能
//					   做基本的获取状态操作，如果要修改路由器状态，需要进一步的授权。
//
// @param interest		收到的命令兴趣包
// @param parameters	命令兴趣包中携带的参数
// @param accept		授权通过则调用此回调
// @param reject		授权拒绝则调用此回调
// @return unc
//
type Authorization func(topPrefix *component.Identifier, interest *packet.Interest,
	parameters *mgmt.ControlParameters,
	accept AuthorizationAccept,
	reject AuthorizationReject) bool

//
// 一个回调函数，用于对收到的控制命令参数进行验证
//
// @Description:
// @param parameters	要验证的参数
// @return bool
//
type ValidateParameters func(parameters *mgmt.ControlParameters) bool

//
// 一个回调函数，用于对收到的已授权的命令进行处理（每个管理模块会通过传入本回调函数自己实现对特定命令的处理逻辑）
//
// @Description:
// @param topPrefix		顶级管理前缀，例如："/min-mir-go/mgmt/localhost"
// @param interest		收到的命令兴趣包
// @param parameters	已通过参数验证的命令参数
// @return *mgmt.ControlResponse	返回一个 ControlResponse 返回给调用方
//
type ControlCommandHandler func(topPrefix *component.Identifier, interest *packet.Interest,
	parameters *mgmt.ControlParameters) *mgmt.ControlResponse

//
// 一个回调函数，处理收到的请求数据集的命令
//
// @Description:
// @param topPrefix
// @param interest
// @param context
//
type StatusDatasetHandler func(topPrefix *component.Identifier, interest *packet.Interest,
	context *StatusDatasetContext)

//
// 兴趣包处理回调
//
// @Description:
// @param topPrefix
// @param interest
//
type InterestHandler func(topPrefix *component.Identifier, interest *packet.Interest)

//
// 定义一个管理命令解析和分发程序
//
// @Description:
//	1. 每个管理模块实现时，需要调用本类进行命令注册；
//	2. 本类负责与MIR进行具体的通信，收到命令兴趣包之后进行解析，解析完分发给对应的管理模块进行处理，管理模块处理完成后，再将处理的结果通过
//	   本类发送给MIR；
//
type IDispatcher interface {
	//
	// 添加一个顶层前缀
	//
	// @Description:
	// @param topPrefix
	//
	AddTopPrefix(topPrefix *component.Identifier)

	//
	// 移除一个顶层前缀
	//
	// @Description:
	// @param topPrefix
	//
	RemoveTopPrefix(topPrefix *component.Identifier)

	//
	// 注册一个控制命令
	//
	// @Description:
	// @param relPrefix				模块名 + 命令名，例如："/fib-mgmt/add"
	// @param authorization			授权验证回调
	// @param validateParameters	参数验证回调
	// @param handler				命令处理回调
	//
	AddControlCommand(relPrefix *component.Identifier, authorization Authorization, validateParameters ValidateParameters,
		handler ControlCommandHandler) error

	//
	// 注册一个数据集
	//
	// @Description:
	//	1. 对于收到的请求数据集的 Interest，首先判断是否携带版本号和分段号，如果携带则直接不响应（因为不带版本号和分段号的请求认为是请求的起始，
	//		会触发数据集收集、分段和发布，携带版本号和分段号的请求认为是对某个具体的数据集的快照的请求，只能期望在缓存中命中，如果没有命中缓存，则应该
	//		重新发起一个不带版本号和序列号的请求，以触发数据集的收集、分段和发布）；
	//	2. 接着就使用 authorization 对请求进行授权鉴定，授权没有通过，则返回错误信息，通过则执行下一步；
	//	3. 然后触发 handler 回调，让对应的管理模块对数据集进行收集、分段和发布。
	// @param relPrefix				模块名 + 命令名，例如："/fib-mgmt/add"
	// @param authorization			授权验证回调
	// @param handler				数据集处理回调
	//
	AddStatusDataset(relPrefix *component.Identifier, authorization Authorization, handler StatusDatasetHandler)

	//
	// 发送一个控制命令应答
	//
	// @Description:
	// @param response	要响应的应答数据
	// @param interest	收到的命令请求兴趣包
	//
	sendControlResponse(response *mgmt.ControlResponse, interest *packet.Interest)

	//
	// 发送一个数据包到 MIR
	//
	// @Description:
	// @param data
	//
	sendData(data *packet.Data)

	//
	// 查询 Dispatcher 内置缓存，尝试获取数据集分段数据
	//
	// @Description:
	// @param topPrefix
	// @param interest
	// @param missStorage
	//
	queryStorage(topPrefix *component.Identifier, interest *packet.Interest, missStorage InterestHandler)
}
