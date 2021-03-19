/**
 * @Author: wzx
 * @Description:
 * @Version: 1.0.0
 * @Date: 2021/3/1 下午11:46
 *@Copyright: MIN-Group；国家重大科技基础设施——未来网络北大实验室；深圳市信息论与未来网络重点实验室
 */

package table

import (
	"fmt"
	"sync"
)

//
// 最长前缀树的表示
// @Description:
//	1.node的表示是一个递归的表示，在最外层公开声明一个结构体屏蔽内部实现细节
//
type LpmMatcher struct {
	node
}

//
// 最长前缀树的节点
//
// @Description:
//	1.最长前缀树中包括本节点的val值，一个子节点列表，和一个用于保护本地val修改的锁
//	2.map的key是子节点对应的前缀(string)，value是node
//	3.val是interface{}的接口实现，调用获取的时候需要手动匹配真实类型
//
type node struct {
	val   interface{}
	table *sync.Map
	lock  *sync.RWMutex
}

//temp,not uesed
type nodeAndLock struct {
	node *node
	lock *sync.RWMutex
}

//
// 初始化,主要是避免每次节点递归的时候都需要判断是否存在lock
//
// @Description:
// @return
//
func (n *node) Create() {
	n.lock = &sync.RWMutex{}
}

//
// 判断节点为空
//
// @Description:节点为空的条件有两个：本节点的val是空的，同时子节点的列表也是空的
// @return bool
//

func (n *node) empty() bool {
	if n.val != nil {
		return false
	}
	if n.table == nil {
		return true
	}

	count := 0
	n.table.Range(func(key, value interface{}) bool {
		count++
		return true
	})

	return count == 0
}

//
// 用于返回，避免val重复的为空判断
//
// @Description:bool为空表示该节点没有val值
// @return (interface{}, bool)
//

func deref(val interface{}) (interface{}, bool) {
	if val == nil {
		return nil, false
	}
	return val, true
}

//
// 最长前缀匹配
//
// @Description:
//   1.逐层去查找查找key数组的第一级是否存在于自己的子节点列表中，直到key值长度是0，表面该节点是查找到的精确匹配值
//   2.如果在递归查找过程中，某个key不存在于该节点的子节点列表中，那么该节点就是最长前缀匹配值
//   3.需要注意的是，在查找路径的每个路径都会加一个读锁
//   4.返回值的bool表示该节点对应的val值是否真实存在
// @return (interface{}, bool)
//

func (n *node) FindLongestPrefixMatch(key []string) (interface{}, bool) {
	n.lock.RLock()
	defer n.lock.RUnlock()
	if len(key) == 0 {
		return deref(n.val)
	}

	if n.table == nil {
		return deref(n.val)
	}
	childInterface, ok := n.table.Load(key[0])
	if !ok {
		return deref(n.val)
	}

	child := childInterface.(*node)
	//child.lock.RLock()
	val, found := child.FindLongestPrefixMatch(key[1:])
	//child.lock.RUnlock()

	return val, found
}

//
// 精确匹配
//
// @Description:
//   1.逐层去查找查找key数组的第一级是否存在于自己的子节点列表中，直到key值长度是0，表面该节点是查找到的精确匹配值
//   2.需要注意的是，在查找路径的每个路径都会加一个读锁
//   3.返回值的bool表示该节点对应的val值是否真实存在
// @return (interface{}, bool)
//

func (n *node) FindExactMatch(key []string) (interface{}, bool) {
	n.lock.RLock()
	defer n.lock.RUnlock()
	if len(key) == 0 {
		return deref(n.val)
	}
	if n.table == nil {

		return deref(nil)
	}
	childInterface, ok := n.table.Load(key[0])
	if !ok {
		return deref(n.val)
	}

	child := childInterface.(*node)
	//child.lock.RLock()
	val, found := child.FindExactMatch(key[1:])
	//child.lock.RUnlock()

	return val, found
}

//
// 添加或者更新
//
// @Description:
//   1.逐层去查找查找key数组的第一级是否存在于自己的子节点列表中，直到key值长度是0，表面该节点是查找到的精确匹配值,那么就更新该节点的值
//   2.如果在某个节点发现该key不存在于子节点列表中，则新建一个节点并加入列表，以此递归到key长度为0,并设置该值
//   3.val值可以为空，回调函数f也可以为空
//   4.在回调函数中通常进行val项的更新
//   5.最后一个节点的修改前需要获取该节点的写锁
// @return (interface{}, bool)
//

func (n *node) AddOrUpdate(key []string, val interface{}, f func(val interface{}) interface{}) interface{} {
	// 写锁
	n.lock.Lock()
	if len(key) == 0 {
		// 下面两条语句应该不会执行
		if val != nil {
			n.val = val
		}
		// 处理函数不空 对val进行处理
		if f != nil {
			n.val = f(&n.val)
		}
		n.lock.Unlock()
		return n.val
	}

	if n.table == nil {
		n.table = new(sync.Map)
	}
	if _, ok := n.table.Load(key[0]); !ok {
		var nal *node
		if len(key) == 1 {
			nal = &node{table: nil, lock: new(sync.RWMutex)}
		} else {
			nal = &node{table: new(sync.Map), lock: new(sync.RWMutex)}
		}

		n.table.Store(key[0], nal)
	}
	n.lock.Unlock()

	n.lock.RLock()
	childInterface, ok := n.table.Load(key[0])
	var tmp interface{}
	if !ok {
		n.lock.RUnlock()
		n.lock.Lock()
		nal := &node{table: new(sync.Map), lock: new(sync.RWMutex)}
		n.table.Store(key[0], nal)
		childInterface, _ = n.table.Load(key[0])
		child := childInterface.(*node)
		//if len(key) > 1 {
		//	//	child.lock.RLock()
		//	child.AddOrUpdate(key[1:], val, f)
		//	//	child.lock.RUnlock()
		//} else {
		//	//	child.lock.Lock()
		//	child.AddOrUpdate(key[1:], val, f)
		//	//	child.lock.Unlock()
		//}
		tmp = child.AddOrUpdate(key[1:], val, f)
		n.lock.Unlock()
	} else {
		child := childInterface.(*node)
		tmp = child.AddOrUpdate(key[1:], val, f)
		n.lock.RUnlock()
	}
	return tmp
}

func (n *node) Delete(key []string) error {
	n.lock.Lock()
	if len(key) == 0 {
		if n.val != nil {
			n.val = nil
			n.lock.Unlock()
			return nil
		}
		n.lock.Unlock()
		return createNodeErrorByType(DataNotExistedError)
	}
	n.lock.Unlock()

	if n.table == nil {
		return createNodeErrorByType(NodeNotExistedError)
	}

	n.lock.RLock()
	childInterface, ok := n.table.Load(key[0])

	if !ok {
		n.lock.RUnlock()
		return createNodeErrorByType(DataNotExistedError)
	}
	child := childInterface.(*node)
	err := child.Delete(key[1:])
	n.lock.RUnlock()

	n.lock.Lock()
	//child.lock.Lock()
	if child.empty() {
		n.table.Delete(key[0])
	}
	//child.lock.Unlock()node traverse
	n.lock.Unlock()
	return err
}

// 适用于遍历所有节点的操作 加入f函数自定义
func (n *node) TraverseFunc(f func(val interface{}) uint64) uint64 {
	var count uint64 = 0

	// table!=nil 说明还有分支
	if n.table != nil {
		n.table.Range(func(key, value interface{}) bool {
			count += value.(*node).TraverseFunc(f)
			return true
		})
	}
	// table == nil 叶子节点
	// val!=nil 说明存有表项  这时候table 可能为空 （叶子节点存储数据） 可能不空 （中间节点存储数据）
	if n.val != nil {
		count += f(n.val)
	}
	return count
}

func (n *node)GetDepth() int{
	var max int = 0
	// table!=nil 说明还有分支 得到最大的分支高度
	if n.table != nil {
		n.table.Range(func(key, value interface{}) bool {
			if tmp:=value.(*node).GetDepth();max<tmp{
				max = tmp
			}
			return true
		})
	}
	return max+1
}


/////////////////////////////////////////////////////////////////////////////////////////////////////////
///// 错误处理
/////////////////////////////////////////////////////////////////////////////////////////////////////////

const (
	NodeNotExistedError = iota
	DataNotExistedError
)

type NodeError struct {
	msg string
}

func (i NodeError) Error() string {
	return fmt.Sprintf("NodeError: %s", i.msg)
}

func createNodeErrorByType(errorType int) (err NodeError) {
	switch errorType {
	case NodeNotExistedError:
		err.msg = "the node is not existed"
	case DataNotExistedError:
		err.msg = "the entry is not existed"
	default:
		err.msg = "Unknown error"
	}
	return
}
