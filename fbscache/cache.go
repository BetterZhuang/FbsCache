/**
    @author: zzg
    @date: 2022/3/14 19:40
    @dir_path:
    @note:
**/

package fbscache

import (
	"fbscache/lru"
	"sync"
)

type cache struct {
	mu sync.Mutex
	lru *lru.Cache
	cacheBytes int64
}

func (c *cache) add(key string, value ByteView)  {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.lru == nil{ //nil时再创建实例，一个对象的延迟初始化意味着该对象的创建将会延迟至第一次使用该对象时。主要用于提高性能，并减少程序内存要求。  -> 单例模式？
		c.lru = lru.New(c.cacheBytes, nil)
	}
	c.lru.Put(key,value)
}

func (c *cache) get(key string) (value ByteView, ok bool)  {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.lru== nil{
		return
	}

	if v,ok := c.lru.Get(key);ok{
		return v.(ByteView),ok
	}
	return
}