/**
    @author: zzg
    @date: 2022/3/14 16:19
    @dir_path: lru
    @note:
**/

package lru

import (
	"fmt"
	"reflect"
	"testing"
)

type String string

func (d String) Len() int {
	return len(d)
}


//test Put
func TestGet(t  *testing.T)  {
	lru := New(int64(0), nil)
	lru.Put("key1", String("1234"))
	if v,ok:=lru.Get("key1"); !ok && string(v.(String)) !="1234"{
		t.Fatalf("cache hit key1=1234 failed!")
	}
	if _, ok := lru.Get("key2");ok{
		t.Fatalf("cache miss key2 failed")
	}
}

//测试，当使用内存超过了设定值时，是否会触发“无用”节点的移除
func TestRemoveOldest(t *testing.T)  {
	k1, k2, k3 := "key1","key2","key3"
	v1,v2,v3 := "val1", "val2","val3"
	cap := len(k1+k2+v1+v2)
	lru := New(int64(cap),nil)
	lru.Put(k1, String(v1))
	fmt.Printf("lru.nbyte: %d\n",lru.nbytes)
	lru.Put(k2, String(v2))
	//fmt.Printf("lru.nbyte: %d\n",lru.nbytes)
	lru.Put(k3, String(v3))
	//fmt.Printf("lru.nbyte: %d\n",lru.nbytes)

	if _, ok := lru.Get("k1"); ok || lru.Len() != 2{
		t.Fatalf("RemoveOldest key1 failed!")
	}
}

//测试回调函数能否被调用
func TestOnEvicted(t *testing.T)  {
	keys := make([]string, 0)
	callback := func(key string, value Value) {
		keys = append(keys, key)
	}
	lru := New(int64(10), callback)
	fmt.Printf("lru.nbyte: %d\n",lru.nbytes)
	fmt.Printf("lru.maxbyte: %d\n",lru.maxBytes)
	lru.Put("key1", String("v12345"))
	fmt.Printf("lru.nbyte: %d\n",lru.nbytes)
	lru.Put("k2", String("v2")) //key1移出
	fmt.Printf("lru.nbyte: %d\n",lru.nbytes)
	lru.Put("k3", String("v3"))
	fmt.Printf("lru.nbyte: %d\n",lru.nbytes)
	lru.Put("k4", String("v4")) //k2移出,v2还在
	fmt.Printf("lru.nbyte: %d\n",lru.nbytes)

	expect := []string{"key1","k2"}
	if !reflect.DeepEqual(expect, keys){
		t.Fatalf("Call OnEvicted failed, expect keys equal to %s", expect)
	}

	//reflect.DeepEqual() 此函数采用两个参数，其值可以是任意类型，即x，y -> 返回布尔值
	//基本类型值等比较会使用==，当比较array、slice的成员、map映射的键值对、struct结构体的字段时，需要用DeepEqual()
	/*深等价:
		x和y同nil或同non-nil
		x和y具有相同的长度
		x和y指向同一个底层数组所初始化的实体对象
	*/
}