//matcher.go文件用于生成和操作对象的字段，以支持过滤、选择等操作
//目前仅仅使用对象的name作为字段，后续可以进行补充

package generic

import (
	"hit.edu/framework/pkg/apimachinery/fields"
	"hit.edu/framework/pkg/apis/meta"
)

// ObjectMetaFieldsSet 返回对象元数据的字段集合，目前仅包含name字段
func ObjectMetaFieldsSet(objectMeta *meta.ObjectMeta) fields.Set {
	return fields.Set{
		"metadata.name": objectMeta.Name,
	}
	
}

// AdObjectMetaField 向字段集合中添加name字段
func AddObjectMetaFieldsSet(source fields.Set, objectMeta *meta.ObjectMeta) fields.Set {
	source["metadata.name"] = objectMeta.Name
	return source
}

// MergeFieldsSets 将一个字段集合中所有字段合并到另一个集合
func MergeFieldsSets(source fields.Set, fragment fields.Set) fields.Set {
	for k, value := range fragment {
		source[k] = value
	}
	return source
}
