// path.go提供了用于验证path的函数
package helper

import (
	"fmt"
	"strings"
)

// NameMayNodeBe定义了不能作为路径的名称
var NameMayNotBe = []string{".", ".."}

// NameMayNotContain定义了不能出现在路径中的符号
var NameMayNotContain = []string{"/", "%"}

// IsValidPathSegmentName 验证路径是否有效
func IsValidPathSegmentName(name string) []string {
	for _, illegalName := range NameMayNotBe {
		if name == illegalName {
			return []string{fmt.Sprintf(`may not be '%s'`, illegalName)}
		}
	}

	var errors []string
	for _, illegalContent := range NameMayNotContain {
		if strings.Contains(name, illegalContent) {
			errors = append(errors, fmt.Sprintf(`may not contain '%s'`, illegalContent))
		}
	}

	return errors
}

// IsValidPathSegmentPrefix 验证路径的前缀是否有效
func IsValidPathSegmentPrefix(name string) []string {
	var errors []string
	for _, illegalContent := range NameMayNotContain {
		if strings.Contains(name, illegalContent) {
			errors = append(errors, fmt.Sprintf(`may not contain '%s'`, illegalContent))
		}
	}

	return errors
}

// ValidatePathSegmentName 根据prefix调用上面两个函数
func ValidatePathSegmentName(name string, prefix bool) []string {
	if prefix {
		return IsValidPathSegmentPrefix(name)
	}

	return IsValidPathSegmentName(name)
}
