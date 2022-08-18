/**
    @author: zzg
    @date: 2022/3/15 21:29
    @dir_path: consistenthash
    @note:
**/

package consistenthash

import (
	"hash/crc32"
	"sort"
	"strconv"
)

type Hash func(data []byte) uint32

type Map struct {
	hash     Hash
	replicas int
	keys []int //Sorted
	hashmap map[int]string
}

func New(replicas int, fn Hash) *Map {
	m := &Map{
		replicas: replicas,
		hash: fn,
		hashmap: make(map[int]string),
	}

	if m.hash == nil{
		m.hash = crc32.ChecksumIEEE
	}
	return m
}

// Add 添加真实节点
func (m *Map)Add (keys ...string)  {
	for _,key := range keys{
		for i:=0;i<m.replicas;i++{
			hash := int(m.hash([]byte(strconv.Itoa(i)+key)))
			m.keys = append(m.keys, hash)
			m.hashmap[hash] = key
		}
	}
	sort.Ints(m.keys)
}

// Get 选择节点
func (m *Map) Get(key string) string {
	if len(m.keys)==0{
		return ""
	}
	hash := int(m.hash([]byte(key)))

	idx := sort.Search(len(m.keys), func(i int) bool {
		return m.keys[i] >= hash
	})

	return m.hashmap[m.keys[idx%len(m.keys)]]
}