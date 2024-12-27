package meta

import (
	"hit.edu/framework/pkg/apimachinery/runtime/schema"
	"net/url"
	"reflect"
	"testing"
)

var paramsCodec = ParameterCodec
var metaGroupVersion = schema.GroupVersion{Group: "meta", Version: "v1"}

func TestConvert_url_Values_To_v1_CreateOptions(t *testing.T) {
	tests := []struct {
		name     string
		input    url.Values
		expected CreateOptions
	}{
		{
			name: "Valid input with all fields",
			input: url.Values{
				"dryRun":          []string{"All"},
				"fieldManager":    []string{"test-manager"},
				"fieldValidation": []string{"Warn"},
			},
			expected: CreateOptions{
				DryRun:          []string{"All"},
				FieldManager:    "test-manager",
				FieldValidation: "Warn",
			},
		},
		{
			name:     "Empty input",
			input:    url.Values{},
			expected: CreateOptions{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := &CreateOptions{}
			_ = paramsCodec.DecodeParameters(tt.input, metaGroupVersion, result)

			// 检查结果是否符合预期
			if !reflect.DeepEqual(*result, tt.expected) {
				t.Errorf("Expected result: %+v, got: %+v", tt.expected, result)
			}
		})
	}
}

func TestConvert_url_Values_To_v1_GetOptions(t *testing.T) {
	tests := []struct {
		name     string
		input    url.Values
		expected GetOptions
	}{
		{
			name: "Valid input with all fields",
			input: url.Values{
				"resourceVersion": []string{"3222"},
			},
			expected: GetOptions{
				ResourceVersion: "3222",
			},
		},
		{
			name:     "Empty input",
			input:    url.Values{},
			expected: GetOptions{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := &GetOptions{}
			_ = paramsCodec.DecodeParameters(tt.input, metaGroupVersion, result)

			// 检查结果是否符合预期
			if !reflect.DeepEqual(*result, tt.expected) {
				t.Errorf("Expected result: %+v, got: %+v", tt.expected, result)
			}
		})
	}
}

func TestConvert_url_Values_To_v1_UpdateOptions(t *testing.T) {
	tests := []struct {
		name     string
		input    url.Values
		expected UpdateOptions
	}{
		{
			name: "Valid input with all fields",
			input: url.Values{
				"dryRun":          []string{"All"},
				"fieldManager":    []string{"test-manager"},
				"fieldValidation": []string{"Warn"},
			},
			expected: UpdateOptions{
				DryRun:          []string{"All"},
				FieldManager:    "test-manager",
				FieldValidation: "Warn",
			},
		},
		{
			name:     "Empty input",
			input:    url.Values{},
			expected: UpdateOptions{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := &UpdateOptions{}
			_ = paramsCodec.DecodeParameters(tt.input, metaGroupVersion, result)

			// 检查结果是否符合预期
			if !reflect.DeepEqual(*result, tt.expected) {
				t.Errorf("Expected result: %+v, got: %+v", tt.expected, result)
			}
		})
	}
}

func TestConvert_url_Values_To_v1_DeleteOptions(t *testing.T) {
	tests := []struct {
		name     string
		input    url.Values
		expected DeleteOptions
	}{
		{
			name: "Valid input with all fields",
			input: url.Values{
				"gracePeriodSeconds": []string{"10"},
				"dryRun":             []string{"All"},
				"orphanDependents":   []string{"true"},
				"propagationPolicy":  []string{"Foreground"},
			},
			expected: DeleteOptions{
				GracePeriodSeconds: int64Ptr(10),
				DryRun:             []string{"All"},
				OrphanDependents:   boolPtr(true),
				PropagationPolicy:  deletionPropagationPtr("Foreground"),
			},
		},
		{
			name:     "Empty input",
			input:    url.Values{},
			expected: DeleteOptions{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := &DeleteOptions{}
			_ = paramsCodec.DecodeParameters(tt.input, metaGroupVersion, result)

			// 检查结果是否符合预期
			if !reflect.DeepEqual(*result, tt.expected) {
				t.Errorf("Expected result: %+v, got: %+v", tt.expected, result)
			}
		})
	}
}

func TestConvert_url_Values_To_v1_ListOptions(t *testing.T) {
	tests := []struct {
		name     string
		input    url.Values
		expected ListOptions
	}{
		{
			name: "Valid input with all fields",
			input: url.Values{
				"labelSelector":       {"key=value"},
				"fieldSelector":       {"field=value"},
				"watch":               {"true"},
				"allowWatchBookmarks": {"false"},
				"resourceVersion":     {"12345"},
				"timeoutSeconds":      {"30"},
				"limit":               {"100"},
				"continue":            {"token"},
				"sendInitialEvents":   {"true"},
			},
			expected: ListOptions{
				LabelSelector:       "key=value",
				FieldSelector:       "field=value",
				Watch:               true,
				AllowWatchBookmarks: false,
				ResourceVersion:     "12345",
				TimeoutSeconds:      int64Ptr(30),
				Limit:               100,
				Continue:            "token",
				SendInitialEvents:   boolPtr(true),
			},
		},
		{
			name:     "Empty input",
			input:    url.Values{},
			expected: ListOptions{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := &ListOptions{}
			_ = paramsCodec.DecodeParameters(tt.input, metaGroupVersion, result)

			// 检查结果是否符合预期
			if !reflect.DeepEqual(*result, tt.expected) {
				t.Errorf("Expected result: %+v, got: %+v", tt.expected, result)
			}
		})
	}
}

func int64Ptr(i int64) *int64 { return &i }
func boolPtr(b bool) *bool    { return &b }
func deletionPropagationPtr(dp string) *DeletionPropagation {
	dpVal := DeletionPropagation(dp)
	return &dpVal
}
