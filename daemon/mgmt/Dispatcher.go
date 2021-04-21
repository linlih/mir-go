// Package mgmt
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
	"github.com/sirupsen/logrus"
	"minlib/common"
	"minlib/component"
	"minlib/encoding"
	"minlib/logicface"
	"minlib/mgmt"
	"minlib/packet"
	"minlib/security"
	common2 "mir-go/daemon/common"
	"mir-go/daemon/lf"
	"mir-go/daemon/table"
	"sync"
)

// Module
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

// Dispatcher
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

// Start
// 调度器启动函数
//
// @Description:启动调度器进行收包监听
//
func (d *Dispatcher) Start() {
	go func() {
		if d.FaceClient == nil {
			common.LogFatal("faceClient is null!")
			return
		}

		// 开始循环处理管理命令兴趣包
		for {
			minPacket, err := d.FaceClient.ReceivePacket(-1)
			if err != nil {
				_ = d.FaceClient.Shutdown()
				common.LogFatal("receive packet fail!the err is:", err)
			}

			// 判断是否是管理包，不是管理包则直接丢弃
			if minPacket.PacketType != encoding.TlvPacketMINManagement {
				continue
			}

			// 创建管理兴趣包
			interest, err := packet.CreateInterestByMINPacket(minPacket)
			if err != nil {
				common.LogWarn("can not parse minPacket to interest!the err is:", err)
				continue
			}

			// 判断命令前缀是否足够长
			prefix := interest.GetName()
			if prefix.Size() < 5 {
				common.LogWarn("Command Interest's prefix size < 5! drop it")
				response := MakeControlResponse(400, "Command Interest's prefix size < 5!", "")
				d.sendControlResponse(response, interest)
				continue
			}

			// 获取顶级前缀，eg: /min-mir/mgmt/localhost
			topPrefix, err := prefix.GetSubIdentifier(0, 4)
			if err != nil {
				common.LogWarnWithFields(logrus.Fields{
					"prefix": prefix.ToUri(),
				}, "Get Command Interest's topPrefix failed!")
				response := MakeControlResponse(400, "Get Command Interest's topPrefix failed!", "")
				d.sendControlResponse(response, interest)
				continue
			}

			// 获取相对前缀，eg: /face-mgmt/list
			relPrefix, err := prefix.GetSubIdentifier(4, 2)
			if err != nil {
				common.LogWarnWithFields(logrus.Fields{
					"prefix": prefix.ToUri(),
				}, "Get Command Interest's relPrefix failed!")
				response := MakeControlResponse(400, "Get Command Interest's relPrefix failed!", "")
				d.sendControlResponse(response, interest)
				continue
			}

			// 获取对应命令的处理回调
			module := d.module[relPrefix.ToUri()]
			if module == nil {
				common.LogWarnWithFields(logrus.Fields{
					"prefix":    prefix.ToUri(),
					"relPrefix": relPrefix.ToUri(),
				}, "the command is not registered!")
				response := MakeControlResponse(400, "the command is not registered!", "")
				d.sendControlResponse(response, interest)
				continue
			}

			// 解析命令参数
			parameters, err := mgmt.ParseControlParameters(interest)
			if err != nil {
				common.LogWarnWithFields(logrus.Fields{
					"interest": interest.ToUri(),
				}, "Parse command interest parameters failed!", err)
			}

			// 进行权限验证
			// TODO: 要求发送管理命令的用户拥有一定级别的权限
			module.authorization(topPrefix, interest, parameters, func() {
				// Accept => 权限验证通过，进行进一步处理

				// 首先查询缓存中有没有匹配项，有则直接返回
				d.queryStorage(topPrefix, interest, func(topPrefix *component.Identifier, interest *packet.Interest) {
					// 如果没用命中缓存，则进一步交给具体的管理模块处理

					// 如果是管理命令，则调用管理命令处理
					if module.ccHandler != nil {
						common.LogDebug("CCHandler")
						if module.validateParameters != nil && !module.validateParameters(parameters) {
							response := MakeControlResponse(400, "Parameters validate failed!", "")
							d.sendControlResponse(response, interest)
							return
						} else {
							response := module.ccHandler(topPrefix, interest, parameters)
							d.sendControlResponse(response, interest)
							return
						}
					} else {
						common.LogWarn("This module doesn't register command handle")
					}

					// 如果是请求数据集的命令，则调用数据集处理回调进行处理
					if module.sdHandler != nil {
						common.LogDebug("SDHandler")
						var context = CreateSDC(interest, d.sendData, d.sendControlResponse, d.saveData)
						module.sdHandler(topPrefix, interest, context)
					}
				})
			}, func(errorType int) {
				// Reject => 权限验证失败，返回错误
				d.sendControlResponse(MakeControlResponse(errorType, "Authorization Failed!", ""), interest)
			})
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
	parameters *component.ControlParameters,
	accept AuthorizationAccept,
	reject AuthorizationReject) {
	if _, ok := d.topPrefixList[topPrefix.ToUri()]; !ok {
		// 顶级域不存在
		reject(0)
		return
	}
	// 没有权限
	if topPrefix.ToUri() == "" {
		reject(1)
		return
	}
	accept()
	return
}

// CreateDispatcher
// 创建调度器函数
//pp
// @Description:创建调度器函数，对调度器进行初始化
//
func CreateDispatcher(config *common2.MIRConfig) *Dispatcher {
	return &Dispatcher{
		topPrefixList: make(map[string]*component.Identifier),
		module:        make(map[string]*Module),
		topLock:       new(sync.RWMutex),
		moduleLock:    new(sync.RWMutex),
		Cache:         New(config.ManagementConfig.CacheSize, nil),
	}
}

// AddTopPrefix
// 添加顶级域函数
//
// @Description:在顶级域map中注册顶级域 顶级域分为本地:/min-mir/mgmt/localhost
// 				远程:/<路由器前缀>/min-mir/mgmt/remote 等
//
func (d *Dispatcher) AddTopPrefix(topPrefix *component.Identifier, fib *table.FIB, serverFace *lf.LogicFace) {
	d.topLock.Lock()
	defer d.topLock.Unlock()
	d.topPrefixList[topPrefix.ToUri()] = topPrefix

	// 在转发表中添加指向管理模块的前缀
	fib.AddOrUpdate(topPrefix, serverFace, 0).SetReadOnly()
}

// RemoveTopPrefix
// 删除顶级域函数
//
// @Description:在map中删除顶级域
//
func (d *Dispatcher) RemoveTopPrefix(topPrefix *component.Identifier) {
	d.topLock.Lock()
	defer d.topLock.Unlock()
	delete(d.topPrefixList, topPrefix.ToUri())
}

// AddControlCommand
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
		ccHandler:          handler,
		sdHandler:          nil,
		missStorage:        nil,
	}
	return nil
}

// AddStatusDataset
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
		relPrefix:          relPrefix,
		authorization:      authorization,
		validateParameters: nil,
		ccHandler:          nil,
		sdHandler:          handler}
	return nil

}

//
// 查询缓存函数
//
// @Description:查询分片缓存，若存在直接取出发送，若不存在从表中请求数据，并分片存储到缓存当中
//
func (d *Dispatcher) queryStorage(topPrefix *component.Identifier, interest *packet.Interest, missStorage InterestHandler) {
	// 如果在缓存中找到分片
	if v, ok := d.Cache.Get(interest.GetName().ToUri()); ok {
		common.LogDebug("hit the cache")
		d.sendData(v.(*packet.Data))
	} else {
		// 没找到 发起请求数据 并添加到缓存中
		common.LogDebug("miss the cache")
		missStorage(topPrefix, interest)
	}
}

//
// 发送控制回复给客户端
//
// @Description:发送控制回复给客户端
//
func (d *Dispatcher) sendControlResponse(response *mgmt.ControlResponse, interest *packet.Interest) {
	if dataByte, err := json.Marshal(response); err == nil {
		data := new(packet.Data)
		data.SetName(interest.GetName())
		data.SetValue(dataByte)
		d.sendData(data)
	} else {
		common.LogError("Marshal data fail!,the err is:", err)
	}
}

// 发送数据包给客户端并缓存数据包
//
// @Description:发送数据包给客户端并缓存数据包
//
func (d *Dispatcher) saveData(data *packet.Data) {
	d.Cache.Add(data.ToUri(), data)
}

// 发送数据包给客户端
//
// @Description:发送数据包给客户端
//
func (d *Dispatcher) sendData(data *packet.Data) {
	// 直接发出的包设置不缓存
	data.NoCache.SetNoCache(true)
	if err := d.FaceClient.SendData(data); err != nil {
		common.LogError("send data fail!the err is :", err)
		return
	}
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
		err.msg = "the top prefixes is empty"
	case CommandExisted:
		err.msg = "the command is already existed"
	default:
		err.msg = "Unknown error"
	}
	return
}
