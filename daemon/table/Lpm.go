/**
 * @Author: wzx
 * @Description:
 * @Version: 1.0.0
 * @Date: 2021/3/1 下午11:46
 *@Copyright: MIN-Group；国家重大科技基础设施——未来网络北大实验室；深圳市信息论与未来网络重点实验室
 */

package table

import (
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

func deref(val interface{}) (interface{}, bool){
	if val == nil {
		return nil, false
	}
	return val, true
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

	child := childInterface.(*node)
	//child.lock.RLock()
	val, found := child.FindLongestPrefixMatch(key[1:])
	//child.lock.RUnlock()

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

	child := childInterface.(*node)
	//child.lock.RLock()
	val, found := child.FindLongestPrefixMatch(key[1:])
	//child.lock.RUnlock()

	return val,found
}

func (n *node) AddOrUpdate(key []string, val interface{}, f func(val interface{}) interface{}) {
	n.lock.Lock()
	if len(key) == 0 {
		if val != nil{
			n.val = val
		}

		if f != nil {
			n.val = f(&n.val)
		}
		n.lock.Unlock()
		return
	}

	if n.table == nil {
		n.table = new(sync.Map)
	}

	if _, ok := n.table.Load(key[0]); !ok {
		nal :=  &node{table: new(sync.Map), lock: new(sync.RWMutex)}
		n.table.Store(key[0], nal)
	}
	n.lock.Unlock()

	n.lock.RLock()
	childInterface, ok := n.table.Load(key[0])
	if !ok{
		n.lock.RUnlock()
		n.lock.Lock()
		nal := &node{table: new(sync.Map), lock: new(sync.RWMutex)}
		n.table.Store(key[0], nal)
		childInterface, _= n.table.Load(key[0])
		child := childInterface.(*node)
		if len(key) > 1{
		//	child.lock.RLock()
			child.AddOrUpdate(key[1:], val, f)
		//	child.lock.RUnlock()
		}else {
		//	child.lock.Lock()
			child.AddOrUpdate(key[1:], val, f)
		//	child.lock.Unlock()
		}
		n.lock.Unlock()
	}else {
		child := childInterface.(*node)
		child.AddOrUpdate(key[1:], val, f)
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

	if n.table == nil {
		return
	}

	n.lock.RLock()
	childInterface, ok := n.table.Load(key[0])

	if !ok{
		n.lock.RUnlock()
		return
	}
	child := childInterface.(*node)

	if len(key) > 1{
	//	child.lock.RLock()
		child.Delete(key[1:])
	//	child.lock.RUnlock()
	}else {
	//	child.lock.Lock()
		child.Delete(key[1:])
	//	child.lock.Unlock()
	}
	n.lock.RUnlock()

	n.lock.Lock()
	//child.lock.Lock()
	if child.empty() {

		n.table.Delete(key[0])
	}
	//child.lock.Unlock()
	n.lock.Unlock()
}
