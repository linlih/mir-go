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

type LpmMatcher struct {
	node
}

type node struct {
	val   interface{}
	table *sync.Map
	lock  *sync.RWMutex
}

type nodeAndLock struct {
	node *node
	lock *sync.RWMutex
}

func (n *node) Create() {
	n.lock = &sync.RWMutex{}
}

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

func deref(val interface{}) (interface{}, bool) {
	if val == nil {
		return nil, false
	}
	return val, true
}

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
