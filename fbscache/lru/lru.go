/**
    @author: zzg
    @date: 2022/3/14 15:48
    @dir_path: lru
    @note:
**/

package lru

import "container/list"

type Cache struct {
	maxBytes int64                    //允许使用的最大内存
	nbytes   int64                    //当前已使用的内存
	ll       *list.List               //双向链表
	cache    map[string]*list.Element //键是字符串，值是双向链表中对应节点的指针
	// optional and executes when an entry is purged
	OnEvicted func(key string, value Value) //某条记录被移除时的回调函数，可以为 nil
}

//键值对
type entry struct {
	key   string
	value Value
}

//在链表中仍保存每个值对应的 key 的好处在于，淘汰队首节点时，需要用 key 从字典中删除对应的映射。

// Value 用Len统计花费多少字节
type Value interface {
	Len() int
}

//为了通用性，我们允许值是实现了 Value 接口的任意类型，该接口只包含了一个方法 Len() int，用于返回值所占用的内存大小。

// New 用来实例化缓存
func New(maxBytes int64, onEvicted func(string, Value)) *Cache {
	return &Cache{
		maxBytes:  maxBytes,
		ll:        list.New(),
		cache:     make(map[string]*list.Element),
		OnEvicted: onEvicted,
	}
}

// Get 查找
func (c *Cache) Get(key string) (value Value, ok bool) {
	if ele, ok := c.cache[key]; ok { //键对应的链表节点存在
		c.ll.MoveToFront(ele)    //节点移到队尾  （队头是最久未使用）
		kv := ele.Value.(*entry) //类型转换（list存储的是任意类型，需要转换为自定义类型）   interface类型转换是.(被转换类型)
		return kv.value, true
	}
	return
}

// RemoveOldest 移除最久未使用节点
func (c *Cache) RemoveOldest() {
	ele := c.ll.Back() //取到队首节点，
	if ele != nil {
		c.ll.Remove(ele) //从链表中删除。
		kv := ele.Value.(*entry)
		delete(c.cache, kv.key)                                //从字典中 c.cache 删除该节点的映射关系
		c.nbytes -= int64(len(kv.key)) + int64(kv.value.Len()) //更新当前所用的内存 c.nbytes。
		if c.OnEvicted != nil {                                //调用回调函数。
			c.OnEvicted(kv.key, kv.value)
		}
	}
}

//已存在 -》 修改 -》移动到到队尾
//不存在 -》添加 -》检查长度 ——》删除最久未使用（队首）

// Put 添加或修改
func (c *Cache) Put(key string, value Value)  {
	if ele,ok := c.cache[key];ok{  //has  ->  change
		c.ll.MoveToFront(ele)
		kv := ele.Value.(*entry)
		c.nbytes += int64(value.Len()) - int64(kv.value.Len())  //已使用空间只需要加上差值
		kv.value = value
	}else {
		ele := c.ll.PushFront(&entry{ //新建节点并加到队尾
			key,
			value,
		})
		c.cache[key] = ele  //字典添加映射关系
		c.nbytes += int64(len(key)) + int64(value.Len())  //更新已用字节长度
	}
	for c.maxBytes != 0 && c.maxBytes < c.nbytes{//如果超过了设定的最大值 c.maxBytes，则移除最少访问的节点。
		c.RemoveOldest()
	}
}

// Len 用来获取添加了多少条数据
func (c *Cache)Len() int {
	return c.ll.Len()
}


