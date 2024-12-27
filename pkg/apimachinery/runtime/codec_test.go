package runtime_test

import (
	"hit.edu/framework/pkg/apimachinery/runtime/schema"
	metav1 "hit.edu/framework/pkg/apis/meta"
	"hit.edu/framework/pkg/client-go/clients/scheme"
	"hit.edu/framework/pkg/component-base/logs"
	"k8s.io/utils/pointer"
	
	"net/url"
	"reflect"
	"testing"
)

func TestEncodeDecodeListOptions(t *testing.T) {
	logs.Init("Test")
	codec := scheme.ParameterCodec
	
	// 测试数据
	originalOptions := &metav1.ListOptions{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ListOptions",
			APIVersion: "v1",
		},
		Watch:             true,
		TimeoutSeconds:    pointer.Int64Ptr(30),
		SendInitialEvents: pointer.BoolPtr(false),
	}
	
	logs.Info("Original options:", *originalOptions)
	// Encode 测试
	queryParams, err := codec.EncodeParameters(originalOptions, schema.GroupVersion{})
	if err != nil {
		t.Fatalf("EncodeParameters failed: %v", err)
	}
	
	expectedParams := url.Values{
		"kind":              []string{"ListOptions"},
		"apiVersion":        []string{"v1"},
		"watch":             []string{"true"},
		"timeoutSeconds":    []string{"30"},
		"sendInitialEvents": []string{"false"},
	}
	logs.Info("Encoded queryParams:", queryParams)
	if !reflect.DeepEqual(queryParams, expectedParams) {
		t.Errorf("Encoded query params do not match expected.\nGot:  %v\nWant: %v", queryParams, expectedParams)
	}
	
	// Decode 测试
	decodedOptions := &metav1.ListOptions{}
	err = codec.DecodeParameters(queryParams, schema.GroupVersion{}, decodedOptions)
	if err != nil {
		t.Fatalf("DecodeParameters failed: %v", err)
	}
	
	logs.Info("Decoded options:", decodedOptions)
	if !reflect.DeepEqual(decodedOptions, originalOptions) {
		t.Errorf("Decoded options do not match original.\nGot:  %+v\nWant: %+v", decodedOptions, originalOptions)
	}
}
