package runtime

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/fxamacker/cbor/v2"
)

// client.go中的UnsupportedMediaType用到
const (
	ContentTypeJSON     string = "application/json"
	ContentTypeYAML     string = "application/yaml"
	ContentTypeProtobuf string = "application/vnd.kubernetes.protobuf"
	ContentTypeCBOR     string = "application/cbor"
)

type TypeMeta struct {
	Kind       string `json:"kind,omitempty"`       // 表示资源类型
	APIVersion string `json:"apiVersion,omitempty"` // 表示 API 版本
}

type Unknown struct {
	TypeMeta `json:",inline" protobuf:"bytes,1,opt,name=typeMeta"`
	// Raw will hold the complete serialized object which couldn't be matched
	// with a registered type. Most likely, nothing should be done with this
	// except for passing it through the system.
	Raw []byte `json:"-" protobuf:"bytes,2,opt,name=raw"`
	// ContentEncoding is encoding used to encode 'Raw' data.
	// Unspecified means no encoding.
	ContentEncoding string `protobuf:"bytes,3,opt,name=contentEncoding"`
	// ContentType  is serialization method used to serialize 'Raw'.
	// Unspecified means ContentTypeJSON.
	ContentType string `protobuf:"bytes,4,opt,name=contentType"`
}

type RawExtension struct {
	// Raw is the underlying serialization of this object.
	//
	// TODO: Determine how to detect ContentType and ContentEncoding of 'Raw' data.
	Raw []byte `json:"-" protobuf:"bytes,1,opt,name=raw"`
	// Object can hold a representation of this extension - useful for working with versioned
	// structs.
	Object Object `json:"-"`
}

// MarshalJSON 使RawExtension支持序列化
func (re RawExtension) MarshalJSON() ([]byte, error) {
	if re.Raw == nil {
		// TODO: this is to support legacy behavior of JSONPrinter and YAMLPrinter, which
		// expect to call json.Marshal on arbitrary versioned objects (even those not in
		// the scheme). pkg/kubectl/resource#AsVersionedObjects and its interaction with
		// kubectl get on objects not in the scheme needs to be updated to ensure that the
		// objects that are not part of the scheme are correctly put into the right form.
		if re.Object != nil {
			return json.Marshal(re.Object)
		}
		return []byte("null"), nil
	}

	contentType := re.guessContentType()
	if contentType == ContentTypeJSON {
		return re.Raw, nil
	}

	u, err := rawToUnstructured(re.Raw, contentType)
	if err != nil {
		return nil, err
	}
	return json.Marshal(u)
}

var (
	cborNull          = []byte{0xf6}
	cborSelfDescribed = []byte{0xd9, 0xd9, 0xf7}
)

func (re RawExtension) guessContentType() string {
	switch {
	case bytes.HasPrefix(re.Raw, cborSelfDescribed):
		return ContentTypeCBOR
	case len(re.Raw) > 0:
		switch re.Raw[0] {
		case '\t', '\r', '\n', ' ', '{', '[', 'n', 't', 'f', '"', '-', '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
			// Prefixes for the four whitespace characters, objects, arrays, strings, numbers, true, false, and null.
			return ContentTypeJSON
		}
	}
	return ""
}

// RawExtension intentionally avoids implementing value.UnstructuredConverter for now because the
// signature of ToUnstructured does not allow returning an error value in cases where the conversion
// is not possible (content type is unrecognized or bytes don't match content type).
func rawToUnstructured(raw []byte, contentType string) (interface{}, error) {
	switch contentType {
	case ContentTypeJSON:
		var u interface{}
		if err := json.Unmarshal(raw, &u); err != nil {
			return nil, fmt.Errorf("failed to parse RawExtension bytes as JSON: %w", err)
		}
		return u, nil
	case ContentTypeCBOR:
		var u interface{}
		if err := cbor.Unmarshal(raw, &u); err != nil {
			return nil, fmt.Errorf("failed to parse RawExtension bytes as CBOR: %w", err)
		}
		return u, nil
	default:
		return nil, fmt.Errorf("cannot convert RawExtension with unrecognized content type to unstructured")
	}
}
