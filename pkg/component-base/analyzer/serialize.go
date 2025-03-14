package analyzer

import (
	"encoding/json"
	"gopkg.in/yaml.v2"
)

/*
SerializeToJson ： 序列化为json文件
接受参数 ： 任意数量的结构体指针（结构体在types.go中定义）
输出： 输出序列化的内容在output.json
*/
func SerializeToJson(v interface{}) (string, error) {
	// 序列化为格式化的 JSON 字符串
	result, err := json.MarshalIndent(v, "", "  ")
	//result, err := json.Marshal(v)
	if err != nil {
		return "", err
	}
	return string(result), err
}

/*
SerializeToYaml ： 序列化为yaml字符串
接受参数 ： 任意数量的结构体（结构体在types.go中定义）
返回： yaml字符串
*/
func SerializeToYaml(v interface{}) (string, error) {
	result, err := yaml.Marshal(v)
	if err != nil {
		return "", err
	}
	return string(result), nil
}
