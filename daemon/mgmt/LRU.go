// Package mgmt
// @Author: yzy
// @Description:
// @Version: 1.0.0
// @Date: 2021/3/10 3:13 下午
// @Copyright: MIN-Group；国家重大科技基础设施——未来网络北大实验室；深圳市信息论与未来网络重点实验室
//
package mgmt

import "container/list"

// Cache
// @Description:分片缓存结构体 包含允许存入的最大包数、当前存入的包数、
//				实现缓存的双链表结构和哈希map结构
// 				最后一个是删除数据的回调函数，暂时保留
//
type Cache struct {
	max       int64                               //允许存入的最大数据包的个数
	count     int64                               //当前已存入的数据包个数
	ll        *list.List                          // go语言自带实现的双向链表
	cache     map[string]*list.Element            // map 一个字符串 对应 一个 链表元素
	OnEvicted func(key string, value interface{}) // 某条记录被移除时的回调函数，可以为 nil
}

// entry
// 缓存中存储的具体内容
//
// @Description:缓存中存储的具体内容，第一个内容索引方便删除数据，第二个参数存储的具体内容
//
type entry struct {
	key   string // 淘汰队首节点时，需要用 key 从字典中删除对应的映射
	value interface{}
}

// New
// 缓存初始化函数
//
// @Description:缓存初始化函数
//
func New(max int64, onEvicted func(string, interface{})) *Cache {
	return &Cache{
		max:       max,
		ll:        list.New(),
		cache:     make(map[string]*list.Element),
		OnEvicted: onEvicted,
	}
}

// Get
// 从缓存中获取数据
//
// @Description:从缓存中获取数据，规则遵循LRU
// @receiver c
// Return:interface{},bool
//
func (c *Cache) Get(key string) (value interface{}, ok bool) {
	if ele, ok := c.cache[key]; ok { //取map
		c.ll.MoveToFront(ele)                 //元素移动到队首
		if kv, ok := ele.Value.(*entry); ok { //断言
			return kv.value, true
		}
	}
	return
}

// RemoveOldest
// 从缓存中删除数据
//
// @Description:从缓存中删除数据，删除的规则遵循LRU
// @receiver c
//
func (c *Cache) RemoveOldest() {
	// 获取front
	if ele := c.ll.Back(); ele != nil {
		c.ll.Remove(ele)
		kv := ele.Value.(*entry)
		// 获取key 从map中删除
		delete(c.cache, kv.key)
		c.count -= 1
		if c.OnEvicted != nil {
			c.OnEvicted(kv.key, kv.value)
		}
	}
}

// Add
// 在缓存中添加数据
//
// @Description:在缓存中添加数据，规则遵循LRU
// @receiver c
//
func (c *Cache) Add(key string, value interface{}) {
	// 如果存在 则修改
	if ele, ok := c.cache[key]; ok {
		// 访问 放到队首
		c.ll.MoveToFront(ele)
		// 取出原entry
		kv := ele.Value.(*entry)
		// 赋值
		kv.value = value
	} else {
		// 不存在则添加 返回节点指针
		ele = c.ll.PushFront(&entry{
			key:   key,
			value: value,
		})
		// 存到map
		c.cache[key] = ele
		// 加内存字节
		c.count += 1
	}
	for c.count > c.max && c.max != 0 {
		c.RemoveOldest()
	}
}

// Len
// 获取缓存中数据包个数
//
// @Description:获取缓存中数据包个数
// @receiver c
//
func (c *Cache) Len() int {
	return c.ll.Len()
}
