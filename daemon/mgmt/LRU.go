//
// @Author: Jianming Que
// @Description:
// @Version: 1.0.0
// @Date: 2021/3/10 3:13 下午
// @Copyright: MIN-Group；国家重大科技基础设施——未来网络北大实验室；深圳市信息论与未来网络重点实验室
//
package mgmt

import "container/list"
//

type Cache struct {
	maxBytes  int64                         //允许使用的最大内存
	nBytes    int64                         //当前已使用的内存
	ll        *list.List                    // go语言自带实现的双向链表
	cache     map[string]*list.Element      // map 一个字符串 对应 一个 链表元素
	OnEvicted func(key string, value Value) // 某条记录被移除时的回调函数，可以为 nil
}
// 存储在链表中的数据
type entry struct {
	data   string // 淘汰队首节点时，需要用 key 从字典中删除对应的映射
	value  Value
}
// 接口 必须继承Len
type Value interface {
	Len() int
}
// 实例化函数
func New(maxBytes int64, onEvicted func(string, Value)) *Cache {
	return &Cache{
		maxBytes:  maxBytes,
		ll:        list.New(),
		cache:     make(map[string]*list.Element),
		OnEvicted: onEvicted,
	}
}
// 查找函数 通过map找到对应的节点元素 取出 value 将该节点元素 放到队列尾
// 对尾 队首 相对 这里约定 front 为队尾
func (c *Cache) Get(key string) (value Value, ok bool) {
	if ele, ok := c.cache[key]; ok { //取map
		c.ll.MoveToFront(ele)                 //元素移动到队首
		if kv, ok := ele.Value.(*entry); ok { //断言
			return kv.value, true
		}
	}
	return
}
// 删除函数 删除最少被访问的节点 队首节点 back 删除节点函数
func (c *Cache) RemoveOldest() {
	// 获取front
	if ele := c.ll.Back(); ele != nil {
		c.ll.Remove(ele)
		kv := ele.Value.(*entry)
		// 获取key 从map中删除
		delete(c.cache, kv.key)
		c.nBytes -= int64(len(kv.key)) + int64(kv.value.Len())
		if c.OnEvicted != nil {
			c.OnEvicted(kv.key, kv.value)
		}
	}
}
// 新增和修改函数
func (c *Cache) Add(key string, value Value) {
	// 如果存在 则修改
	if ele, ok := c.cache[key]; ok {
		// 访问 放到队首
		c.ll.MoveToFront(ele)
		// 取出原entry
		kv := ele.Value.(*entry)
		// 减去原值value len 加上 修改值 value len
		c.nBytes += int64(value.Len() - kv.value.Len())
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
		c.nBytes += int64(len(key) + value.Len())
	}
	for c.nBytes > c.maxBytes && c.maxBytes != 0 {
		c.RemoveOldest()
	}
}
func (c *Cache) Len() int {
	return c.ll.Len()
}
