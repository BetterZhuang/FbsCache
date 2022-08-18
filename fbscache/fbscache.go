/**
    @author: zzg
    @date: 2022/3/14 19:49
    @dir_path:
    @note:
**/

package fbscache

import (
	pb "fbscache/fbscachepb"
	"fbscache/singleflight"
	"fmt"
	"log"
	"sync"
)

type Getter interface {
	Get(key string) ([]byte, error)
}

type GetterFunc func(key string) ([]byte, error)

func (f GetterFunc) Get(key string)([]byte , error)  {
	return f(key)
}


//Group 是 FbsCache 最核心的数据结构，负责与用户的交互，并且控制缓存值存储和获取的流程。
//一个Group看做一个缓存空间
type Group struct {
	name      string
	getter    Getter //缓存未命中时获取源数据的回调(callback)
	mainCache cache

	//day5
	peers PeerPicker

	//day6
	loader *singleflight.Group
}

var (
	mu sync.Mutex
	groups = make(map[string]*Group)
)

// NewGroup 创建Group实例, 并且将 group 存储在全局变量 groups 中
func NewGroup(name string, cacheBytes int64, getter Getter) *Group {
	if getter == nil{
		panic("nil getter")
	}
	mu.Lock()
	defer mu.Unlock()
	g := &Group{
		name:      name,
		getter:    getter,
		mainCache: cache{cacheBytes: cacheBytes},
		loader:    &singleflight.Group{}, //day6 add
	}
	groups[name] = g
	return g
}

//GetGroup 用来特定名称的 Group，这里使用了只读锁 RLock()
func GetGroup(name string) *Group {
	mu.Lock()
	g := groups[name]
	mu.Unlock()
	return g
}

func (g *Group) Get(key string) (ByteView, error)  {
	if key ==""{
		return ByteView{}, fmt.Errorf("key is required")
	}

	if v,ok := g.mainCache.get(key);ok{
		log.Println("[FbsCache] hit")
		return v,nil
	}
	return g.load(key)
}


//func (g *Group) load(key string)(value ByteView, err error)  {
//	return g.getlocally(key)
//}

func (g *Group) getlocally(key string)(ByteView, error)  { //单机场景，从本地获取
	bytes, err := g.getter.Get(key)  //调用用户回调函数 g.getter.Get() 获取源数据，并且将源数据添加到缓存 mainCache 中（通过 populateCache 方法）
	if err != nil {
		return ByteView{}, err
	}
	value := ByteView{b: cloneBytes(bytes)}
	g.populateCache(key ,value)
	return value,nil
}

func (g *Group)populateCache(key string, value ByteView)  {
	g.mainCache.add(key, value)
}

func (g *Group) RegisterPeers(peers PeerPicker)  {
	if g.peers != nil{
		panic("RegisterPeerPicker called more than once")
	}
	g.peers = peers
}

func (g *Group) load(key string) (value ByteView, err error)  {
	// each key is only fetched once (either locally or remotely)
	// regardless of the number of concurrent callers.

	viewi, err := g.loader.Do(key, func() (interface{}, error) {
		if g.peers != nil {
			if peer, ok := g.peers.PickPeer(key); ok {
				if value, err = g.getFromPeer(peer, key); err == nil {
					return value, nil
				}
				log.Println("[fbscache failed to get from peers", err)
			}
		}
		return g.getlocally(key)
	})

	if err == nil{
		return viewi.(ByteView), nil
	}
	return
}

func (g *Group) getFromPeer(peer PeerGetter, key string) (ByteView, error) {
	//bytes, err := peer.Get(g.name, key)
	//if err != nil {
	//	return ByteView{}, err
	//}
	//return ByteView{b: bytes}, err

	//day7
	req := &pb.Request{
		Group: g.name,
		Key: key,
	}
	res := &pb.Response{}
	err := peer.Get(req,res)
	if err != nil {
		return ByteView{}, err
	}
	return ByteView{b:res.Value}, err
}