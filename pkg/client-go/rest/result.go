package rest

import (
	"fmt"
	"hit.edu/framework/pkg/apimachinery/runtime"
	metav1 "hit.edu/framework/pkg/apis/meta"
	"hit.edu/framework/pkg/component-base/logs"
	"k8s.io/apimachinery/pkg/util/net"
	"mime"
)

// TODO: 替换K8s的模块

// Results
// Result contains the result of calling Request.Do().
type Result struct {
	body        []byte
	warnings    []net.WarningHeader
	contentType string
	err         error
	statusCode  int

	decoder runtime.Decoder
}

func (r Result) Into(obj runtime.Object) error {
	// TODO: 格式匹配
	// 填充Object
	if r.err != nil {
		// Check whether the result has a Status object in the body and prefer that.
		return r.Error()
	}

	mediaType, _, err := mime.ParseMediaType(r.contentType)
	// Check if contentType is "text/plain"
	if mediaType == "text/plain" {
		return nil
	}

	if r.decoder == nil {
		return fmt.Errorf("serializer for %s doesn't exist", r.contentType)
	}
	if len(r.body) == 0 {
		return fmt.Errorf("0-length response with status code: %d and content type: %s",
			r.statusCode, r.contentType)
	}
	out, _, err := r.decoder.Decode(r.body, nil, obj)
	if err != nil {
		return err
	}
	switch t := out.(type) {
	case *metav1.Status:
		// any status besides StatusSuccess is considered an error.
		//除StatusSuccess之外的任何状态都被视为错误。

		if t.Status != metav1.StatusSuccess {
			logs.Info("出现了除StatusSuccess之外的状态")
			//todo: 除StatusSuccess之外的任何状态 报错处理
			//return errors.FromObject(t)
		}
	}
	return nil
}
