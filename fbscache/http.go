/**
    @author: zzg
    @date: 2022/3/14 21:31
    @dir_path:
    @note:
**/

package fbscache

import (
	"fbscache/consistenthash"
	pb "fbscache/fbscachepb"
	"fmt"
	"github.com/golang/protobuf/proto"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"
	"sync"
)

const (
	defaultbasepath = "/_fbscache/"
	defaultReplicas = 50
)


// HTTPPool implements
type HTTPPool struct {
	self string  //自身地址：主机名/IP+端口
	basePath string  //节点间通讯地址的前缀

	//new add
	mu sync.Mutex
	peers *consistenthash.Map
	httpGetters map[string]*httpGetter
}

func NewHTTPPool(self string) *HTTPPool {
	return &HTTPPool{
		self:     self,
		basePath: defaultbasepath,
	}
}

func (p *HTTPPool) Log(format string, v ...interface{})  {
	log.Printf("[Server %s] %s", p.self, fmt.Sprintf(format, v...))
}

func (p *HTTPPool) ServeHTTP(w http.ResponseWriter, req *http.Request)  {
	if !strings.HasPrefix(req.URL.Path, p.basePath){  //判断访问路径的前缀是否是 basePath
		panic("HTTPPool serving unexpected path: " + req.URL.Path)
	}
	p.Log("%s %s", req.Method, req.URL.Path)
	// /<basepath>/<groupname>/<key> required
	parts := strings.SplitN(req.URL.Path[len(p.basePath):],"/",2)
	if len(parts)!=2{
		http.Error(w, "bad request",http.StatusBadRequest)
		return
	}

	groupName := parts[0]
	key := parts[1]

	group := GetGroup(groupName) //通过 groupname 得到 group 实例
	if group == nil{
		http.Error(w, "no such group:"+groupName, http.StatusNotFound)
		return
	}

	view, err := group.Get(key) //使用 group.Get(key) 获取缓存数据
	//if err != nil {
	//	http.Error(w, err.Error(), http.StatusInternalServerError)
	//	return
	//}
	//
	//w.Header().Set("Content-Type","applicaton/octet-stream")
	//w.Write(view.ByteSlice()) //使用 w.Write() 将缓存值作为 httpResponse 的 body 返回

	body, err := proto.Marshal(&pb.Response{Value: view.ByteSlice()})
	if err != nil {
		http.Error(w,err.Error(),http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type","applicaton/octet-stream")
	w.Write(body)


}



var _ PeerPicker = (*HTTPPool)(nil)
//day5

type httpGetter struct {
	baseURL string
}

//func (h *httpGetter) Get(group string, key string) ([]byte, error)  {
//	u := fmt.Sprintf("%v%v/%v", h.baseURL, url.QueryEscape(group), url.QueryEscape(key))
//
//	res, err := http.Get(u)
//
//	if err != nil{
//		return nil, err
//	}
//
//	defer res.Body.Close()
//
//	if res.StatusCode != http.StatusOK{
//		return nil, fmt.Errorf("server returned: %v", res.Status)
//	}
//
//	bytes, err := ioutil.ReadAll(res.Body)
//	if err != nil {
//		return nil, fmt.Errorf("reading response body: %v", err)
//	}
//
//	return bytes, nil
//}

func (h *httpGetter) Get(in *pb.Request, out *pb.Response) error {
	u := fmt.Sprintf(
		"%v%v/%v",
		h.baseURL,
		url.QueryEscape(in.GetGroup()),
		url.QueryEscape(in.GetKey()),
	)
	res, err := http.Get(u)
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK{
		return fmt.Errorf("server returned: %v", res.Status)
	}
	bytes, err := ioutil.ReadAll(res.Body)
	if err = proto.Unmarshal(bytes, out); err != nil {
		return fmt.Errorf("decoding response body: %v", err)
	}
	return nil
}

var _ PeerGetter = (*httpGetter)(nil)


func (p *HTTPPool) Set(peers ...string)  {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.peers = consistenthash.New(defaultReplicas, nil)
	p.peers.Add(peers...)
	p.httpGetters = make(map[string]*httpGetter, len(peers))

	for _, peer := range peers{
		p.httpGetters[peer] = &httpGetter{
			baseURL: peer + p.basePath,
		}
	}
}

func (p *HTTPPool) PickPeer(key string)(PeerGetter, bool)  {
	p.mu.Lock()
	defer p.mu.Unlock()
	if peer := p.peers.Get(key); peer != ""&&peer != p.self{
		p.Log("Pick peer %s", peer)
		return p.httpGetters[peer], true
	}
	return nil, false
}










