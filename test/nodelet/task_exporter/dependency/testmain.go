package main

import (
	"fmt"
	"strconv"
	"strings"
)

// Person 是最外层的结构体
type Person struct {
	Name      string
	Age       int
	Addresses map[string]Address // 包含多个 Address 结构体，以 map 形式存储
}

// Address 是中间层的结构体
type Address struct {
	City    string
	Country string
	Phones  []Phone // 包含多个 Phone 结构体
}

// Phone 是最里层的结构体
type Phone struct {
	Type   string
	Number string
}

func main() {
	// 创建一个 Person 实例
	person := Person{
		Name: "Alice",
		Age:  30,
		Addresses: map[string]Address{
			"home": {
				City:    "Beijing",
				Country: "China",
				Phones: []Phone{
					{Type: "home", Number: "123-456-7890"},
					{Type: "work", Number: "098-765-4321"},
				},
			},
			"office": {
				City:    "New York",
				Country: "USA",
				Phones: []Phone{
					{Type: "mobile", Number: "555-123-4567"},
				},
			},
		},
	}

	// 定义要访问的字段路径
	fieldPath := "Addresses{office}.Phones{0}.Number{}"

	// 解析字段路径并访问对应的字段
	value := getValueByPath(person, fieldPath)
	fmt.Println("Value:", value)
}

// getValueByPath 解析字段路径并返回对应的值
func getValueByPath(person Person, path string) interface{} {
	// 按照 '.' 分割路径
	parts := strings.Split(path, ".")

	// 逐步解析路径
	var current interface{}
	current = person
	for _, part := range parts {
		// 解析字段名和索引
		field := strings.Split(part, "{")
		fieldName := field[0]
		var index string
		if len(field) > 1 {
			index = strings.TrimSuffix(field[1], "}")
		} else {
			index = ""
		}

		switch v := current.(type) {
		case Person:
			if fieldName == "Addresses" {
				if v.Addresses != nil {
					if address, ok := v.Addresses[index]; ok {
						current = address
					} else {
						return nil
					}
				} else {
					return nil
				}
			} else {
				return nil
			}
		case Address:
			if fieldName == "Phones" {
				i, err := strconv.Atoi(index)
				if err != nil || i < 0 || i >= len(v.Phones) {
					return nil
				}
				current = v.Phones[i]
			} else {
				return nil
			}
		case Phone:
			if fieldName == "Type" {
				current = v.Type
			} else if fieldName == "Number" {
				current = v.Number
			} else {
				return nil
			}
		default:
			return nil
		}
	}

	// 返回最终的值
	return current
}

