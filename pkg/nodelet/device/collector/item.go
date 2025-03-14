package collector

import (
	apis "hit.edu/framework/pkg/apis/cores"
	"strings"
)

// Item 是数据存储的基本单元
type Item struct {
	// 收集数据的名称
	name string //格式：name = namespace + '.' + subsystem + '.' + name

	// 收集数据的描述信息
	desc string

	// 需要收集的数据的key列表
	labels []string

	// TODO:需要收集的数据的Value
	values map[string]apis.Property //每个key对应的value列表

	err error
}

// NewItem 创造新的Item labels []string --> map[string]string
func NewItem(name string, desc string, labels []string) *Item {
	i := &Item{
		name: name,
		desc: desc,
		// TODO: Labels
		labels: labels, // 初始化 key
		// TODO: Values
		values: make(map[string]apis.Property), // 初始化空的 values map
	}
	return i
}

// UpdateValues 用于动态更新值
func (i *Item) UpdateValues(newValues map[string]apis.Property) {
	for k, v := range newValues {
		i.values[k] = v
	}
}

// NewName 构造收集数据的名称
// name = namespace + '.' + subsystem + '.' + name
func NewName(namespace string, subsystem string, name string) string {
	if name == "" {
		return ""
	}
	// TODO：格式检查
	return strings.Join([]string{namespace, subsystem, name}, ".")
}

// toString 将Item中的values成员转为 key=value,key=value样式的字符串
func (i *Item) toString() string {
	// TODO: 实现toString函数
	str := ""
	//for k, v := range i.values {
	//	str += k + "=" + v + ","
	//}
	return str
}
