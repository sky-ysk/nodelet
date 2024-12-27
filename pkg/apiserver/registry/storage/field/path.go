// path.go定义了与Path相关的操作，用于存储路径
package field

import (
	"bytes"
	"fmt"
	"strconv"
)

type pathOptions struct {
	path *Path
}

// PathOption modifies a pathOptions
type PathOption func(o *pathOptions)

// WithPath生成PathOption
func WithPath(p *Path) PathOption {
	return func(o *pathOptions) {
		o.path = p
	}
}

// ToPath 根据传入的opts生成Path
func ToPath(opts ...PathOption) *Path {
	c := &pathOptions{}
	for _, opt := range opts {
		opt(c)
	}
	return c.path
}

// Path represents the path from some root to a particular field.
type Path struct {
	name   string // the name of this field or "" if this is an index
	index  string // if name == "", this is a subscript (index or map key) of the previous element
	parent *Path  // nil if this is the root element
}

// NewPath 创建Path对象
func NewPath(name string, moreNames ...string) *Path {
	r := &Path{name: name, parent: nil}
	for _, anotherName := range moreNames {
		r = &Path{name: anotherName, parent: r}
	}
	return r
}

// Root 返回Path的根元素
func (p *Path) Root() *Path {
	for ; p.parent != nil; p = p.parent {
		// Do nothing.
	}
	return p
}

// Child 创建当前路径的子路径
func (p *Path) Child(name string, moreNames ...string) *Path {
	r := NewPath(name, moreNames...)
	r.Root().parent = p
	return r
}

// Index 为当前的路径设置整数索引
func (p *Path) Index(index int) *Path {
	return &Path{index: strconv.Itoa(index), parent: p}
}

// Key 将当前路径设置键映射
func (p *Path) Key(key string) *Path {
	return &Path{index: key, parent: p}
}

// String 返回Path对象的字符串表示
func (p *Path) String() string {
	if p == nil {
		return "<nil>"
	}
	elems := []*Path{}
	for ; p != nil; p = p.parent {
		elems = append(elems, p)
	}
	buf := bytes.NewBuffer(nil)
	for i := range elems {
		p := elems[len(elems)-1-i]
		if p.parent != nil && len(p.name) > 0 {
			buf.WriteString(".")
		}
		if len(p.name) > 0 {
			buf.WriteString(p.name)
		} else {
			fmt.Fprintf(buf, "[%s]", p.index)
		}
	}
	return buf.String()
}
