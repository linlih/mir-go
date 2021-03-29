package mgmt

import (
	"fmt"
	"minlib/component"
	"minlib/mgmt"
	"minlib/packet"
	"minlib/security"
	"sync"
)

type Module struct {
	relPrefix          *component.Identifier
	validateParameters ValidateParameters
	authorization      Authorization
	ccHandler          ControlCommandHandler
	sdHandler          StatusDatasetHandler
	missStorage        InterestHandler
}

type Dispacher struct {
	topPrefixList map[string]*component.Identifier // 已经注册的顶级域前缀 map实现 方便取
	module        map[string]*Module
	topLock       *sync.RWMutex // 读写锁
	moduleLock    *sync.RWMutex
	KeyChain      *security.KeyChain       // 网络包签名和验签 发送数据包的时候使用
	SignInfo      *component.SignatureInfo // 表示签名的元数据
	Cache         *Cache
}

func CreateDispatcher() *Dispacher {
	return &Dispacher{
		topPrefixList: make(map[string]*component.Identifier),
		module:        make(map[string]*Module),
		topLock:       new(sync.RWMutex),
		moduleLock:    new(sync.RWMutex),
		Cache:         New(100, nil),
	}
}

// 顶级域加入到切片中
func (d *Dispacher) AddTopPrefix(topPrefix *component.Identifier) {
	d.topLock.Lock()
	defer d.topLock.Unlock()
	d.topPrefixList[topPrefix.ToUri()] = topPrefix
}

// 从切片中删除顶级域
func (d *Dispacher) RemoveTopPrefix(topPrefix *component.Identifier) {
	d.topLock.Lock()
	defer d.topLock.Unlock()
	delete(d.topPrefixList, topPrefix.ToUri())
}

//	在调度器中添加控制命令 add/delete ...
func (d *Dispacher) AddControlCommand(relPrefix *component.Identifier, authorization Authorization, validateParameters ValidateParameters,
	handler ControlCommandHandler) error {
	if len(d.topPrefixList) == 0 {
		return createDispacherErrorByType(TopPrefixsEmpty)
	}
	if _, ok := d.module[relPrefix.ToUri()]; ok {
		return createDispacherErrorByType(CommandExisted)
	}
	moduleLock.Lock()
	defer moduleLock.Unlock()
	d.module[relPrefix.ToUri()] = &Module{
		relPrefix:          relPrefix,
		authorization:      authorization,
		validateParameters: validateParameters,
		ccHandler:          handler}

	return nil
}

// list...命令
func (d *Dispacher) AddStatusDataset(relPrefix *component.Identifier, authorization Authorization, handler StatusDatasetHandler) error {
	// 顶级域空 返回错误
	if len(d.topPrefixList) == 0 {
		return createDispacherErrorByType(TopPrefixsEmpty)
	}
	if _, ok := d.module[relPrefix.ToUri()]; ok {
		fmt.Println("the command is existed!")
		return createDispacherErrorByType(CommandExisted)
	}

	moduleLock.Lock()
	defer moduleLock.Unlock()
	d.module[relPrefix.ToUri()] = &Module{
		relPrefix:     relPrefix,
		authorization: authorization,
		sdHandler:     handler}
	return nil

}

// 查找数据包切片 并发送
func (d *Dispacher) queryStorage(topPrefix *component.Identifier, interest *packet.Interest, missStorage InterestHandler) {
	// 如果在缓存中找到分片
	if v, ok := d.Cache.Get(interest.ToUri()); ok {
		// 发送分片
		d.sendData(v.(*packet.Data))
	} else {
		// 没找到 发起请求数据 并添加到缓存中
		missStorage(topPrefix, interest)
	}
}

func (d *Dispacher) sendControlResponse(response *mgmt.ControlResponse, interest *packet.Interest) {

}

func (d *Dispacher) sendData(data *packet.Data) {

}

/////////////////////////////////////////////////////////////////////////////////////////////////////////
///// 错误处理
/////////////////////////////////////////////////////////////////////////////////////////////////////////

const (
	NotMatchTopPrefix = iota
	TopPrefixsEmpty
	CommandExisted
)

type DispacherError struct {
	msg string
}

func (d DispacherError) Error() string {
	return fmt.Sprintf("DispacherError: %s", d.msg)
}

func createDispacherErrorByType(errorType int) (err DispacherError) {
	switch errorType {
	case NotMatchTopPrefix:
		err.msg = "the command prefix not match top prefix"
	case TopPrefixsEmpty:
		err.msg = "the top prefixs is empty"
	case CommandExisted:
		err.msg = "the command is already existed"
	default:
		err.msg = "Unknown error"
	}
	return
}
