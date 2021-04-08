//
// @Author: yzy
// @Description:
// @Version: 1.0.0
// @Date: 2021/3/10 3:13 下午
// @Copyright: MIN-Group；国家重大科技基础设施——未来网络北大实验室；深圳市信息论与未来网络重点实验室
//
package mgmt

import (
	"fmt"
	"minlib/component"
	"minlib/encoding"
	"minlib/logicface"
	"minlib/mgmt"
	"minlib/packet"
	"minlib/security"
	"mir-go/daemon/common"
	"os"
	"sync"
)

//
// 行为模块结构体
//
// @Description:行为模块结构体，一个行为对应一个模块，如add、delete、list等
//   			同时携带了行为对应的前缀relPrefix 参数验证函数、授权函数、
//				行为对应的注册函数、分片函数、没有命中缓存执行的回调函数
//
type Module struct {
	relPrefix          *component.Identifier // 行为模块前缀 如/fib-mgmt/add
	validateParameters ValidateParameters    //	参数验证函数
	authorization      Authorization         // 授权函数
	ccHandler          ControlCommandHandler // 注册命令回调函数
	sdHandler          StatusDatasetHandler  // 数据处理分片回调函数
	missStorage        InterestHandler       // 没有命中缓存的回调函数
}

//
// 调度器结构体
//
// @Description:调度器结构体，全局变量定义在Init.go中，包含顶级域map、行为模块map、
//   			读写锁，对网络包进行签名和验签、签名元数据、缓存
//
type Dispatcher struct {
	FaceClient    *logicface.LogicFace             // 内部face，用来和转发器进行通信
	topPrefixList map[string]*component.Identifier // 已经注册的顶级域前缀 map实现 方便取 存储前缀如:/min-mir/mgmt/localhost
	module        map[string]*Module               // 行为模块
	topLock       *sync.RWMutex                    // 顶级域map读写锁
	moduleLock    *sync.RWMutex                    // 行为模块map读写锁
	KeyChain      *security.KeyChain               // 网络包签名和验签 发送数据包的时候使用
	SignInfo      *component.SignatureInfo         // 表示签名的元数据
	Cache         *Cache                           // 存储数据包分片缓存
}

//
// 调度器启动函数
//
// @Description:启动调度器进行收包监听
//
func (d *Dispatcher) Start() {
	go func() {
		for {
			if d.FaceClient == nil {
				common.LogError("faceClient is null!")
				os.Exit(0)
			}
			minPacket, err := d.FaceClient.ReceivePacket()
			if err != nil {
				common.LogError("receive packet fail!the err is:", err)
				d.FaceClient.Shutdown()
				os.Exit(0)
			}
			if minPacket.PacketType != encoding.TlvPacketMINCommon {
				//common.LogWarn("receive minPacket from tcp type error")
				//continue
			}
			interest, err := packet.CreateInterestByMINPacket(minPacket)
			if err != nil {
				common.LogError("can not parse minPacket to interest!the err is:", err)
				continue
			}
			actionPrefix, _ := interest.GetName().GetSubIdentifier(3, 2)
			topPrefix, _ := interest.GetName().GetSubIdentifier(0, 3)
			module := d.module[actionPrefix.ToUri()]
			if module == nil {
				common.LogWarn("the command is not registered!")
				continue
			}
			parameters := &mgmt.ControlParameters{}
			if module.authorization(topPrefix, interest, parameters, authorizationAccept, authorizationReject) {

				if module.ccHandler != nil {
					if err := parameters.Parse(interest); err != nil {
						common.LogError("解析控制参数错误！the err is:", err)
						continue
					}
					if !module.validateParameters(parameters) {
						common.LogWarn("parameters validate fail!discard the packet!")
						continue

					}
					module.ccHandler(topPrefix, interest, parameters)
				}

				if module.sdHandler != nil {
					d.queryStorage(topPrefix, interest, func(topPrefix *component.Identifier, interest *packet.Interest) {
						var context = CreateSDC(interest, d.sendDataAndSave, d.sendControlResponse)
						module.sdHandler(topPrefix, interest, context)
					})
				}
			}

		}
	}()
}

//
// 授权验证函数
//
// @Description:对收到的兴趣包中的参数进行解析，并验证权限
// @Return:bool
//
func (d *Dispatcher) authorization(topPrefix *component.Identifier, interest *packet.Interest,
	parameters *mgmt.ControlParameters,
	accept AuthorizationAccept,
	reject AuthorizationReject) bool {
	if _, ok := d.topPrefixList[topPrefix.ToUri()]; !ok {
		// 顶级域不存在
		reject(5)
		return false
	}
	// 没有权限
	if topPrefix.ToUri() == "" {
		reject(6)
		return false
	}

	accept()
	return true
}

//
// 创建调度器函数
//
// @Description:创建调度器函数，对调度器进行初始化
//
func CreateDispatcher() *Dispatcher {
	return &Dispatcher{
		topPrefixList: make(map[string]*component.Identifier),
		module:        make(map[string]*Module),
		topLock:       new(sync.RWMutex),
		moduleLock:    new(sync.RWMutex),
		Cache:         New(100, nil),
	}
}

//
// 添加顶级域函数
//
// @Description:在顶级域map中注册顶级域 顶级域分为本地:/min-mir/mgmt/localhost
// 				远程:/<路由器前缀>/min-mir/mgmt/remote 等
//
func (d *Dispatcher) AddTopPrefix(topPrefix *component.Identifier) {
	d.topLock.Lock()
	defer d.topLock.Unlock()
	d.topPrefixList[topPrefix.ToUri()] = topPrefix
}

//
// 删除顶级域函数
//
// @Description:在map中删除顶级域
//
func (d *Dispatcher) RemoveTopPrefix(topPrefix *component.Identifier) {
	d.topLock.Lock()
	defer d.topLock.Unlock()
	delete(d.topPrefixList, topPrefix.ToUri())
}

//
// 注册控制命令函数
//
// @Description:注册控制命令函数,如:add、delete等控制命令
// @Return:error
func (d *Dispatcher) AddControlCommand(relPrefix *component.Identifier, authorization Authorization, validateParameters ValidateParameters,
	handler ControlCommandHandler) error {
	if len(d.topPrefixList) == 0 {
		return createDispatcherErrorByType(TopPrefixesEmpty)
	}
	moduleLock.RLock()
	if _, ok := d.module[relPrefix.ToUri()]; ok {
		return createDispatcherErrorByType(CommandExisted)
	}
	moduleLock.RUnlock()

	moduleLock.Lock()
	defer moduleLock.Unlock()
	d.module[relPrefix.ToUri()] = &Module{
		relPrefix:          relPrefix,
		authorization:      authorization,
		validateParameters: validateParameters,
		ccHandler:          handler}
	return nil
}

//
// 注册数据集命令函数
//
// @Description:注册数据集命令函数,如:list 等数据集命令
// @Return:error
//
func (d *Dispatcher) AddStatusDataset(relPrefix *component.Identifier, authorization Authorization, handler StatusDatasetHandler) error {
	// 顶级域空 返回错误
	if len(d.topPrefixList) == 0 {
		return createDispatcherErrorByType(TopPrefixesEmpty)
	}
	moduleLock.RLock()
	if _, ok := d.module[relPrefix.ToUri()]; ok {
		common.LogError("the command is existed!")
		return createDispatcherErrorByType(CommandExisted)
	}
	moduleLock.RUnlock()

	moduleLock.Lock()
	defer moduleLock.Unlock()
	d.module[relPrefix.ToUri()] = &Module{
		relPrefix:     relPrefix,
		authorization: authorization,
		sdHandler:     handler}
	return nil

}

//
// 查询缓存函数
//
// @Description:查询分片缓存，若存在直接取出发送，若不存在从表中请求数据，并分片存储到缓存当中
//
func (d *Dispatcher) queryStorage(topPrefix *component.Identifier, interest *packet.Interest, missStorage InterestHandler) {
	// 如果在缓存中找到分片
	if v, ok := d.Cache.Get(interest.ToUri()); ok {
		// 发送分片
		d.sendData(v.(*packet.Data))
	} else {
		// 没找到 发起请求数据 并添加到缓存中
		missStorage(topPrefix, interest)
	}
}

//	TODO:暂未实现
// 发送控制回复给客户端
//
// @Description:发送控制回复给客户端
//
func (d *Dispatcher) sendControlResponse(response *mgmt.ControlResponse, interest *packet.Interest) {

}

// 发送数据包给客户端并缓存数据包
//
// @Description:发送数据包给客户端并缓存数据包
//
func (d *Dispatcher) sendDataAndSave(data *Data) {
	d.Cache.Add(data.key, data.dataFrag)
	// TODO
}

// 发送数据包给客户端
//
// @Description:发送数据包给客户端
//
func (d *Dispatcher) sendData(data *packet.Data) {
	// TODO
}

/////////////////////////////////////////////////////////////////////////////////////////////////////////
///// 错误处理
/////////////////////////////////////////////////////////////////////////////////////////////////////////

const (
	NotMatchTopPrefix = iota
	TopPrefixesEmpty
	CommandExisted
)

type DispatcherError struct {
	msg string
}

func (d DispatcherError) Error() string {
	return fmt.Sprintf("DispatcherError: %s", d.msg)
}

func createDispatcherErrorByType(errorType int) (err DispatcherError) {
	switch errorType {
	case NotMatchTopPrefix:
		err.msg = "the command prefix not match top prefix"
	case TopPrefixesEmpty:
		err.msg = "the top prefixs is empty"
	case CommandExisted:
		err.msg = "the command is already existed"
	default:
		err.msg = "Unknown error"
	}
	return
}
