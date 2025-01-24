package utils

import (
	apis "hit.edu/framework/pkg/apis/cores"
	"hit.edu/framework/pkg/component-base/logs"
)

type ParamsWorker interface {
	ConstructParam(devices map[string]apis.Device, runtime *apis.Runtime) error
	StringToPropertyType(s string) apis.PropertyType
}

// ConstructParam 构造device的参数，存放到ExpectedProperties成员中
func ConstructParam(devices map[string]apis.Device, runtime *apis.Runtime) error {
	// 只有一个device的情况
	if len(devices) == 1 {
		for name, device := range devices {
			logs.Infof("constructing device[%s] param.....\n", name)
			for _, input := range runtime.Inputs {
				property := apis.Property{Value: input.Value, Type: StringToPropertyType(input.ValueType), Name: input.Name}
				device.Spec.ExpectedProperties[input.Name] = property
			}
			logs.Infof("construct device[%s] param done\n", name)
		}
		return nil
	} else { // TODO: 多个device的情况
		paramsNumber := GetParamsNumber(runtime.Image)
		for name, device := range devices {
			logs.Infof("constructing device[%s] param.....\n", name)
			for i := 0; i < len(runtime.Inputs); i += paramsNumber {
				// paramsNumber为一组
				for j := 0; j < paramsNumber && i+j < len(runtime.Inputs); j++ {
					input := runtime.Inputs[i+j]
					property := apis.Property{
						Value: input.Value,
						Type:  StringToPropertyType(input.ValueType),
						Name:  input.Name,
					}
					device.Spec.ExpectedProperties[input.Name] = property
				}
			}
			logs.Infof("construct device[%s] param done\n", name)
		}
	}
	return nil
}

/*
	在填写device的ExpectedProperties属性时，遇到类型不匹配问题。
	ExpectedProperties有自己单独的属性PropertyType
	而input中的ValueType是string类型
	所以根据ValueType的内容进行映射
*/

// StringToPropertyType 用来进行string类型到apis.PropertyType的映射
func StringToPropertyType(s string) apis.PropertyType {
	switch s {
	case "String":
		return apis.StringType
	case "string":
		return apis.StringType
	case "Integer":
		return apis.IntegerType
	case "integer":
		return apis.IntegerType
	case "int":
		return apis.IntegerType
	case "float":
		return apis.DoubleType
	case "Double":
		return apis.DoubleType
	case "URL":
		return apis.URLType
	case "url":
		return apis.URLType
	case "Compose":
		return apis.ComposeType
	case "Boolean":
		return apis.BoolType
	case "boolean":
		return apis.BoolType
	case "bool":
		return apis.BoolType
	case "Bool":
		return apis.BoolType
	default:
		return ""
	}
}

// GetParamsNumber 根据image获得参数的个数
func GetParamsNumber(image string) int {
	switch image {
	case "Move":
		return 3
	case "Grab":
		return 8
	default:
		return 0
	}
}
