package analyzer

import (
	"encoding/json"
	"fmt"
	validator "github.com/xeipuuv/gojsonschema"
	resource "hit.edu/framework/api/resource/v1"
	"hit.edu/framework/pkg/component-base/logs"
	"reflect"
)

/*
Deserialize:    反序列化函数
接受参数： 文件路径 ， 对象接口
功能：
 1. 验证文件格式
 2. 将字符串序列化到对应的
*/
func Deserialize(str string, obj interface{}) (interface{}, error) {
	// 通过反射获取obj的类型
	t := reflect.TypeOf(obj)
	logs.Debugf("Type Of Object %s", t.Name())

	// 先到列表中找到对应的Schema是否支持
	item, ok := resource.Has(t.Name())

	// 如果支持，则进行格式化校验
	// 从Schema列表中找到对应的Schema文件，进行格式验证
	if ok {
		var path string
		// 根据Schema类型，确定读取路径
		if item.Type == resource.HTTP {
			path = item.Schema
		} else {
			path = "file://" + item.Schema
		}
		schema := validator.NewReferenceLoader(path)
		source := validator.NewStringLoader(str)
		result, err := validator.Validate(schema, source)

		if err != nil {
			logs.Errorf("Failed To Validate String %v, Reason is %v", result, err.Error())
			return nil, fmt.Errorf("Failed To Validate String %v, Reason is %v", result, err.Error())
		}

		if result.Valid() {
			logs.Infof("The document is valid\n")
		} else {
			logs.Errorf("The document is not valid. see errors :\n")
			for _, desc := range result.Errors() {
				logs.Errorf("- %s\n", desc)
			}
			return nil, fmt.Errorf("The document is not valid\n")
		}

	} else {
		logs.Errorf("String %s Type is %s, Unsupported Resource Type", str, t.Name())
		return nil, fmt.Errorf("String %s Type is %s, Unsupported Resource Type", str, t.Name())
	}
	// 进行反序列化
	err := json.Unmarshal([]byte(str), &obj)
	if err != nil {
		logs.Errorf("Failed to Deserialize String %v, Reason is %v", str, err.Error())
		return nil, err
	}

	logs.Debugf("Deserialize String %v, result is %v", str, obj)
	return obj, nil
}
