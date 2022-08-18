/**
    @author: zzg
    @date: 2022/3/14 20:06
    @dir_path:
    @note:
**/

package fbscache

import (
	"fmt"
	"log"
	"testing"
)

var db1 = map[string]string{  //模拟本地数据库
	"Tom":"610",
	"mike":"789",
	"jecy":"963",
	"sam":"4286",
}

func TestGet(t *testing.T)  {
	loadCounts := make(map[string]int, len(db1))
	fbs := NewGroup("scores", 2<<10, GetterFunc(
		func(key string) ([]byte, error) {
			log.Println("[slowdb] search key", key)
			if v,ok := db1[key];ok{
				if _,ok:=loadCounts[key]; !ok{
					loadCounts[key] = 0
				}
				loadCounts[key] += 1  //统计每个键调用回调函数的次数
				return []byte(v),nil
			}
			log.Println("loadCounts", loadCounts)
			return nil, fmt.Errorf("%s not exist", key)
		}))


	for k,v := range db1 {
		if view, err := fbs.Get(k);err != nil || view.String() != v{
			t.Fatalf("failed to get value of Tom")
		} //load from callback func
		if _, err := fbs.Get(k);err != nil || loadCounts[k]>1{
			t.Fatalf("cache %s miss", k)
		}  //cache hit
	}

	if view, err := fbs.Get("unknown");err==nil{
		t.Fatalf("the value of unknown should be empty, but %s got", view)
	}
}
