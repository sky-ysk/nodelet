package yaml

import (
	"encoding/json"
	"gopkg.in/yaml.v3"
)

// TODO: Yaml序列化
type Serializer struct {
}

// YAMLToJSON 将 YAML 数据转换为 JSON 格式。
// 它首先解析 YAML 数据，然后将其转为 JSON 格式返回。
// 如果发生错误，它会返回一个错误。
func YAMLToJSON(data []byte) ([]byte, error) {
	// 首先，解析 YAML 数据到一个通用的 map 类型
	var v interface{}
	err := yaml.Unmarshal(data, &v) //go的库
	if err != nil {
		return nil, err // 如果 YAML 解析失败，返回错误
	}

	// 然后，将解析后的数据编码为 JSON 格式
	jsonData, err := json.Marshal(v)
	if err != nil {
		return nil, err // 如果 JSON 编码失败，返回错误
	}

	// 返回 JSON 格式的数据
	return jsonData, nil
}

// JSONToYAML 将 JSON 数据转换为 YAML 格式。
// 它首先解析 JSON 数据为 Go 数据结构，然后将其转换为 YAML 格式。
func JSONToYAML(data []byte) ([]byte, error) {
	// 解析 JSON 数据到 Go 的数据结构
	var jsonData interface{}
	err := json.Unmarshal(data, &jsonData)
	if err != nil {
		return nil, err // 如果 JSON 解码失败，返回错误
	}

	// 将 Go 数据结构转换为 YAML 格式
	yamlData, err := yaml.Marshal(jsonData)
	if err != nil {
		return nil, err // 如果 YAML 编码失败，返回错误
	}

	return yamlData, nil
}
