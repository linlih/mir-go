/**
 * @Author: wzx
 * @Description:
 * @Version: 1.0.0
 * @Date: 2021/3/1 下午11:46
 *@Copyright: MIN-Group；国家重大科技基础设施——未来网络北大实验室；深圳市信息论与未来网络重点实验室
 */

package table

import "sync"

type LpmMatcher struct {
	node
}

type node struct {
	val   *interface{}
	table *sync.Map
	lock  *sync.RWMutex
}

type nodeAndLock struct {
	node   *node
	lock   *sync.RWMutex
}

func (n* node) Create(){
	n.lock = &sync.RWMutex{}
}

func (n *node) empty() bool {
	if n.val != nil {return false}
	if n.table == nil {return true}

	count := 0
	n.table.Range(func(key, value interface{}) bool {
		count++
		return true
	})

	return count == 0
}

func deref(val *interface{}) (interface{}, bool){
	if val == nil {
		return nil, false
	}
	return *val, true
}

func (n *node) FindLongestPrefixMatch(key []string) ( interface{},  bool) {
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

	child := childInterface.(nodeAndLock)
	child.lock.RLock()
	val, found := child.node.FindLongestPrefixMatch(key[1:])
	child.lock.RUnlock()

	return val,found
}

func (n *node) FindExactMatch(key []string) ( interface{},  bool) {
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

	child := childInterface.(nodeAndLock)
	child.lock.RLock()
	val, found := child.node.FindLongestPrefixMatch(key[1:])
	child.lock.RUnlock()

	return val,found
}

func (n *node) AddOrUpdate(key []string, val interface{}) {

	n.lock.Lock()
	if len(key) == 0 {
		n.val = &val
		n.lock.Unlock()
		return
	}

	if n.table == nil {
		n.table = new(sync.Map)
	}

	if _, ok := n.table.Load(key[0]); !ok {
		nal := nodeAndLock{node: &node{table: new(sync.Map), lock: new(sync.RWMutex)}, lock: &sync.RWMutex{}}
		n.table.Store(key[0], nal)
	}
	n.lock.Unlock()

	n.lock.RLock()
	childInterface, ok := n.table.Load(key[0])
	if !ok{
		n.lock.RUnlock()
		n.lock.Lock()
		nal := nodeAndLock{node: &node{table: new(sync.Map), lock: new(sync.RWMutex)}, lock: &sync.RWMutex{}}
		n.table.Store(key[0], nal)
		childInterface, _= n.table.Load(key[0])
		child := childInterface.(nodeAndLock)
		if len(key) > 1{
			child.lock.RLock()
			child.node.AddOrUpdate(key[1:], val)
			child.lock.RUnlock()
		}else {
			child.lock.Lock()
			child.node.AddOrUpdate(key[1:], val)
			child.lock.Unlock()
		}
		n.lock.Unlock()
	}else {
		child := childInterface.(nodeAndLock)
		child.node.AddOrUpdate(key[1:], val)
		n.lock.RUnlock()
	}

}

func (n *node) Delete(key []string) {
	n.lock.Lock()
	if len(key) == 0 {
		n.val = nil
		n.lock.Unlock()
		return
	}
	n.lock.Unlock()

	n.lock.RLock()
	if n.table == nil {
		n.lock.RUnlock()
		return
	}

	childInterface, ok := n.table.Load(key[0])

	if !ok{
		n.lock.RUnlock()
		return
	}
	child := childInterface.(nodeAndLock)

	if len(key) > 1{
		child.lock.RLock()
		child.node.Delete(key[1:])
		child.lock.RUnlock()
	}else {
		child.lock.Lock()
		child.node.Delete(key[1:])
		child.lock.Unlock()
	}
	n.lock.RUnlock()

	n.lock.Lock()
	child.lock.Lock()
	if child.node.empty() {
		n.table.Delete(key[0])
	}
	child.lock.Unlock()
	n.lock.Unlock()
}
