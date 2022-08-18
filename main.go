/**
    @author: zzg
    @date: 2022/3/14 21:47
    @dir_path:
    @note:
**/

package main

import (
	"fbscache"
	"flag"
	"fmt"
	"log"
	"net/http"
)

var db = map[string]string{
	"Tom":"610",
	"mike":"789",
	"jecy":"963",
	"sam":"4286",
}

func createGroup() *fbscache.Group {
	return fbscache.NewGroup("scores", 2<<10, fbscache.GetterFunc(
		func(key string) ([]byte, error) {
			log.Println("[SlowDB] search key ", key)
			if v,ok := db[key];ok{
				return []byte(v),nil
			}
			return nil, fmt.Errorf("%s not exist", key)
		}))
}

func startCacheServer(addr string, addrs []string, fbs *fbscache.Group)  {
	peers := fbscache.NewHTTPPool(addr)
	peers.Set(addrs...)
	fbs.RegisterPeers(peers)
	log.Println("TheCache is running at", addr)
	log.Fatal(http.ListenAndServe(addr[7:], peers))
}

func startAPIServer(apiAddr string, fbs *fbscache.Group)  {
	http.Handle("/api", http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			key := r.URL.Query().Get("key")
			view, err := fbs.Get(key)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Type", "application/octet-stream")
			w.Write(view.ByteSlice())
		}))
	log.Println("API server is running at ", apiAddr)
	log.Fatal(http.ListenAndServe(apiAddr[7:], nil))
}

func main()  {
	//fbscache.NewGroup("scores", 2<<10, fbscache.GetterFunc(
	//	func(key string) ([]byte, error) {
	//		log.Println("[slowdb] search key", key)
	//		if v,ok := db[key];ok{
	//			return []byte(v),nil
	//		}
	//		return nil, fmt.Errorf("%s not exist.", key)
	//	}))
	//
	//addr := "localhost:9999"
	//peers := fbscache.NewHTTPPool(addr)
	//log.Println("fbscache is running at", addr)
	//log.Fatal(http.ListenAndServe(addr, peers))

	//需要命令行传入 port 和 api 2 个参数，用来在指定端口启动 HTTP 服务。
	var port int
	var api bool
	flag.IntVar(&port, "port", 8001, "Cache server port")
	flag.BoolVar(&api, "api", false, "Start a api server?")
	flag.Parse()

	apiAddr := "http://localhost:9999"
	addrMap := map[int]string{
		8001: "http://localhost:8001",
		8002: "http://localhost:8002",
		8003: "http://localhost:8003",
	}

	var addrs []string
	for _, v := range addrMap {
		addrs = append(addrs, v)
	}

	fbs := createGroup()
	if api {
		//启动一个 API 服务（端口 9999），与用户进行交互，用户感知
		go startAPIServer(apiAddr, fbs)
	}
	//启动缓存服务器：创建 HTTPPool，添加节点信息，注册到 gee 中，启动 HTTP 服务（共3个端口，8001/8002/8003），用户不感知
	startCacheServer(addrMap[port], addrs, fbs)
}
