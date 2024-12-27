package responsewriters

import (
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"hit.edu/framework/pkg/apis/meta"
	negotiation "hit.edu/framework/pkg/apiserver/endpoints/handler/negotitation"
	"io"
	"net/http"
	"strconv"
	"strings"
	"sync"
	
	"hit.edu/framework/pkg/apimachinery/runtime"
	"hit.edu/framework/pkg/apimachinery/runtime/schema"
)

// Avoid 发出看起来像有效HTML的错误。引用是可以的。
var sanitizer = strings.NewReplacer(`&`, "&amp;", `<`, "&lt;", `>`, "&gt;")

// SerializeObject 将对象序列化为响应内容并写入 http.ResponseWriter
func SerializeObject(mediaType string, encoder runtime.Encoder, hw http.ResponseWriter, req *http.Request, statusCode int, object runtime.Object) {
	ctx := req.Context()
	req = req.WithContext(ctx)
	
	w := &deferredResponseWriter{
		mediaType:       mediaType,
		statusCode:      statusCode,
		contentEncoding: negotiateContentEncoding(req),
		hw:              hw,
		ctx:             ctx,
	}
	
	err := encoder.Encode(object, w)
	if err == nil {
		err = w.Close()
		if err != nil {
			//logs.Error("apiserver was unable to close cleanly the response writer", zap.String("error:", err.Error()))
		}
		return
	}
	
	status := ErrorToAPIStatus(err)
	candidateStatusCode := int(status.Code)
	// 如果当前状态码成功，则允许错误的状态码覆盖它
	if statusCode >= http.StatusOK && statusCode < http.StatusBadRequest {
		w.statusCode = candidateStatusCode
	}
	output, err := runtime.Encode(encoder, status)
	if err != nil {
		w.mediaType = "text/plain"
		output = []byte(fmt.Sprintf("%s: %s", status.Reason, status.Message))
	}
	if _, err := w.Write(output); err != nil {
		//logs.Error("apiserver was unable to write a fallback JSON response", zap.String("error:", err.Error()))
	}
	w.Close()
}

var gzipPool = &sync.Pool{
	New: func() interface{} {
		gw, err := gzip.NewWriterLevel(nil, defaultGzipContentEncodingLevel)
		if err != nil {
			panic(err)
		}
		return gw
	},
}

const (
	// defaultGzipContentEncodingLevel 默认gzip压缩比
	defaultGzipContentEncodingLevel = 1
	// defaultGzipThresholdBytes 超出此值才会进行gzip压缩
	defaultGzipThresholdBytes = 128 * 1024
)

// negotiateContentEncoding 检查客户端是否支持 gzip 编码。如果支持，返回"gzip"，否则返回空字符串
func negotiateContentEncoding(req *http.Request) string {
	encoding := req.Header.Get("Accept-Encoding")
	if len(encoding) == 0 {
		return ""
	}
	for len(encoding) > 0 {
		var token string
		if next := strings.Index(encoding, ","); next != -1 {
			token = encoding[:next]
			encoding = encoding[next+1:]
		} else {
			token = encoding
			encoding = ""
		}
		switch strings.TrimSpace(token) {
		case "gzip":
			return "gzip"
		}
	}
	return ""
}

type deferredResponseWriter struct {
	mediaType       string
	statusCode      int
	contentEncoding string
	
	hasWritten bool
	hw         http.ResponseWriter
	w          io.Writer
	
	ctx context.Context
}

func (w *deferredResponseWriter) Write(p []byte) (n int, err error) {
	
	if w.hasWritten {
		return w.w.Write(p)
	}
	w.hasWritten = true
	
	hw := w.hw
	header := hw.Header()
	switch {
	case w.contentEncoding == "gzip" && len(p) > defaultGzipThresholdBytes:
		header.Set("Content-Encoding", "gzip")
		header.Add("Vary", "Accept-Encoding")
		
		gw := gzipPool.Get().(*gzip.Writer)
		gw.Reset(hw)
		
		w.w = gw
	default:
		w.w = hw
	}
	
	header.Set("Content-Type", w.mediaType)
	hw.WriteHeader(w.statusCode)
	return w.w.Write(p)
}

func (w *deferredResponseWriter) Close() error {
	if !w.hasWritten {
		return nil
	}
	var err error
	switch t := w.w.(type) {
	case *gzip.Writer:
		err = t.Close()
		t.Reset(nil)
		gzipPool.Put(t)
	}
	return err
}

// WriteObjectNegotiated 以客户端协商的内容类型呈现对象。
func WriteObjectNegotiated(s runtime.NegotiatedSerializer, restrictions negotiation.EndpointRestrictions, gv schema.GroupVersion, w http.ResponseWriter, req *http.Request, statusCode int, object runtime.Object, listGVKInContentType bool) {
	mediaType, serializer, err := negotiation.NegotiateOutputMediaType(req, s, restrictions)
	if err != nil {
		if statusCode < http.StatusOK || statusCode >= http.StatusBadRequest {
			WriteRawJSON(int(statusCode), object, w)
			return
		}
		status := ErrorToAPIStatus(err)
		WriteRawJSON(int(status.Code), status, w)
		return
	}
	if _, ok := object.(*meta.Status); ok {
		gv = schema.GroupVersion{Group: "meta", Version: "v1"}
	}
	encoder := s.EncoderForVersion(serializer.Serializer, gv)
	if listGVKInContentType {
		SerializeObject(generateMediaTypeWithGVK(serializer.MediaType, mediaType.Convert), encoder, w, req, statusCode, object)
	} else {
		SerializeObject(serializer.MediaType, encoder, w, req, statusCode, object)
	}
}

func generateMediaTypeWithGVK(mediaType string, gvk *schema.GroupVersionKind) string {
	if gvk == nil {
		return mediaType
	}
	if gvk.Group != "" {
		mediaType += ";g=" + gvk.Group
	}
	if gvk.Version != "" {
		mediaType += ";v=" + gvk.Version
	}
	if gvk.Kind != "" {
		mediaType += ";as=" + gvk.Kind
	}
	return mediaType
}

// WriteRawJSON 直接序列化为json到响应中
func WriteRawJSON(statusCode int, object interface{}, w http.ResponseWriter) {
	output, err := json.MarshalIndent(object, "", "  ")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	w.Write(output)
}

// ErrorNegotiated 向响应呈现错误
func ErrorNegotiated(err error, s runtime.NegotiatedSerializer, gv schema.GroupVersion, w http.ResponseWriter, req *http.Request) int {
	status := ErrorToAPIStatus(err)
	code := int(status.Code)
	// 当写入错误时，检查状态是否指示在一段时间后重试
	if status.Details != nil && status.Details.RetryAfterSeconds > 0 {
		delay := strconv.Itoa(int(status.Details.RetryAfterSeconds))
		w.Header().Set("Retry-After", delay)
	}
	
	if code == http.StatusNoContent {
		w.WriteHeader(code)
		return code
	}
	
	WriteObjectNegotiated(s, negotiation.DefaultEndpointRestrictions, gv, w, req, code, status, false)
	return code
}

// InternalError 呈现一个简单的内部错误
func InternalError(w http.ResponseWriter, req *http.Request, err error) {
	http.Error(w, sanitizer.Replace(fmt.Sprintf("Internal Server Error: %q: %v", req.RequestURI, err)),
		http.StatusInternalServerError)
}
