/**
    @author: zzg
    @date: 2022/3/14 19:32
    @dir_path:
    @note:
**/

package fbscache

type ByteView struct {  //只读，用来表示缓存值
	b []byte  //存储真实的缓存值，byte支持任意类型（图片、字符串等）
}

func (v ByteView) Len() int { //缓存对象必须实现 Value 接口，即 Len() int 方法，返回其所占的内存大小
	return len(v.b)
}

func (v ByteView) ByteSlice() []byte { //返回一个拷贝，防止缓存值被外部环境修改
	return cloneBytes(v.b)
}

func (v ByteView) String() string {
	return string(v.b)
}
func cloneBytes(b []byte) []byte {
	c := make([]byte, len(b))
	copy(c, b)
	return c
}