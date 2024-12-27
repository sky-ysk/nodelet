package runtime

import (
	"bytes"
	"fmt"
	"hit.edu/framework/pkg/apimachinery/conversion/queryparams"
	"hit.edu/framework/pkg/apimachinery/runtime/schema"
	"net/url"
	"reflect"
	"strings"
)

// codec binds an encoder and decoder.
type codec struct {
	Encoder
	Decoder
}

// NewCodec creates a Codec from an Encoder and Decoder.
func NewCodec(e Encoder, d Decoder) Codec {
	return codec{e, d}
}

// Encode is a convenience wrapper for encoding to a []byte from an Encoder
func Encode(e Encoder, obj Object) ([]byte, error) {
	buf := &bytes.Buffer{}
	if err := e.Encode(obj, buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// Decode is a convenience wrapper for decoding data into an Object.
func Decode(d Decoder, data []byte) (Object, error) {
	obj, _, err := d.Decode(data, nil, nil)
	return obj, err
}

// DecodeInto performs a Decode into the provided object.
func DecodeInto(d Decoder, data []byte, into Object) error {
	out, gvk, err := d.Decode(data, nil, into)
	if err != nil {
		return err
	}
	if out != into {
		return fmt.Errorf("unable to decode %s into %v", gvk, reflect.TypeOf(into))
	}
	return nil
}

type parameterCodec struct {
	typer     ObjectTyper
	convertor ObjectConvertor
	creator   ObjectCreater
	defaulter ObjectDefaulter
}

var _ ParameterCodec = &parameterCodec{}

func NewParameterCodec(scheme *Scheme) ParameterCodec {
	return &parameterCodec{
		typer:     scheme,
		convertor: scheme,
		creator:   scheme,
		defaulter: scheme,
	}
}

// DecodeParameters 将 URL 查询参数解码为目标对象
func (c *parameterCodec) DecodeParameters(parameters url.Values, from schema.GroupVersion, into Object) error {
	// 确保 obj 是一个非空的指针
	val := reflect.ValueOf(into)
	if val.Kind() != reflect.Pointer || val.IsNil() {
		return fmt.Errorf("object must be a non-nil pointer")
	}

	// 获取结构体值
	val = val.Elem()
	if val.Kind() != reflect.Struct {
		return fmt.Errorf("object must point to a struct")
	}

	//通过converter来解码参数
	metaGroupVersion := schema.GroupVersion{Group: "meta", Version: "v1"}
	if from == metaGroupVersion {
		if len(parameters) == 0 {
			return nil
		}
		targetGVKs, _, err := c.typer.ObjectKinds(into)
		if err != nil {
			return err
		}
		for i := range targetGVKs {
			if targetGVKs[i].GroupVersion() == from {
				if err := c.convertor.Convert(&parameters, into, nil); err != nil {
					return err
				}
				// in the case where we going into the same object we're receiving, default on the outbound object
				if c.defaulter != nil {
					c.defaulter.Default(into)
				}
				return nil
			}
		}

		input, err := c.creator.New(from.WithKind(targetGVKs[0].Kind))
		if err != nil {
			return err
		}
		if err := c.convertor.Convert(&parameters, input, nil); err != nil {
			return err
		}
		// if we have defaulter, default the input before converting to output
		if c.defaulter != nil {
			c.defaulter.Default(input)
		}
		return c.convertor.Convert(input, into, nil)
	}

	return decodeStruct(parameters, val)
}

func decodeStruct(parameters url.Values, val reflect.Value) error {
	typ := val.Type()
	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)
		tag, _, inline := parseJSONTag(field)

		// 处理嵌套结构体
		fieldValue := val.Field(i)
		if inline {
			if err := decodeStruct(parameters, fieldValue); err != nil {
				return err
			}
			continue
		}

		// 从查询参数中获取值
		values, exists := parameters[tag]
		if !exists || len(values) == 0 {
			continue
		}

		// 检查字段是否可设置
		if !fieldValue.CanSet() {
			continue
		}

		// 处理指针类型字段
		fieldType := field.Type
		if fieldType.Kind() == reflect.Pointer {
			// 初始化指针
			fieldValue.Set(reflect.New(fieldType.Elem()))
			fieldValue = fieldValue.Elem()
		}

		// 将值写入字段
		if err := queryparams.SetValueFromString(fieldValue, values[0]); err != nil {
			return fmt.Errorf("failed to set field %s: %v", field.Name, err)
		}
	}
	return nil
}

func parseJSONTag(field reflect.StructField) (tag string, omitempty, inline bool) {
	tag = field.Tag.Get("json")
	if tag == "-" {
		return "", false, false
	}
	parts := strings.Split(tag, ",")
	tag = parts[0]
	for _, part := range parts[1:] {
		if part == "omitempty" {
			omitempty = true
		} else if part == "inline" {
			inline = true
		}
	}
	return tag, omitempty, inline
}

// EncodeParameters 将对象编码为 URL 查询参数
func (c *parameterCodec) EncodeParameters(obj Object, to schema.GroupVersion) (url.Values, error) {
	//todo: 实现GroupVersionKind的转化
	// 将对象的字段转化为查询参数
	return queryparams.Convert(obj)
}

// SerializerInfoForMediaType returns the first info in types that has a matching media type (which cannot
// include media-type parameters), or the first info with an empty media type, or false if no type matches.
// SerializerInfoForMediaType返回具有匹配媒体类型（不能包含媒体类型参数）的类型中的第一个信息，
// 或者返回具有空媒体类型的第一个消息，如果没有匹配的类型，则返回false。
func SerializerInfoForMediaType(types []SerializerInfo, mediaType string) (SerializerInfo, bool) {
	for _, info := range types {
		if info.MediaType == mediaType {
			return info, true
		}
	}
	for _, info := range types {
		if len(info.MediaType) == 0 {
			return info, true
		}
	}
	return SerializerInfo{}, false
}

// UseOrCreateObject returns obj if the canonical ObjectKind returned by the provided typer matches gvk, or
// invokes the ObjectCreator to instantiate a new gvk. Returns an error if the typer cannot find the object.
func UseOrCreateObject(t ObjectTyper, c ObjectCreater, gvk schema.GroupVersionKind, obj Object) (Object, error) {
	if obj != nil {
		kinds, _, err := t.ObjectKinds(obj)
		if err != nil {
			return nil, err
		}
		for _, kind := range kinds {
			if gvk == kind {
				return obj, nil
			}
		}
	}
	return c.New(gvk)
}

//// From K8s
//var (
//	// InternalGroupVersioner will always prefer the internal version for a given group version kind.
//	InternalGroupVersioner GroupVersioner = internalGroupVersioner{}
//	// DisabledGroupVersioner will reject all kinds passed to it.
//	DisabledGroupVersioner GroupVersioner = disabledGroupVersioner{}
//)
//
//const (
//	internalGroupVersionerIdentifier = "internal"
//	disabledGroupVersionerIdentifier = "disabled"
//)
//
//type internalGroupVersioner struct{}
//
//// KindForGroupVersionKinds returns an internal Kind if one is found, or converts the first provided kind to the internal version.
//func (internalGroupVersioner) KindForGroupVersionKinds(kinds []schema.GroupVersionKind) (schema.GroupVersionKind, bool) {
//	for _, kind := range kinds {
//		if kind.Version == APIVersionInternal {
//			return kind, true
//		}
//	}
//	for _, kind := range kinds {
//		return schema.GroupVersionKind{Group: kind.Group, Version: APIVersionInternal, Kind: kind.Kind}, true
//	}
//	return schema.GroupVersionKind{}, false
//}
//
//// Identifier implements GroupVersioner interface.
//func (internalGroupVersioner) Identifier() string {
//	return internalGroupVersionerIdentifier
//}
//
//type disabledGroupVersioner struct{}
//
//// KindForGroupVersionKinds returns false for any input.
//func (disabledGroupVersioner) KindForGroupVersionKinds(kinds []schema.GroupVersionKind) (schema.GroupVersionKind, bool) {
//	return schema.GroupVersionKind{}, false
//}
//
//// Identifier implements GroupVersioner interface.
//func (disabledGroupVersioner) Identifier() string {
//	return disabledGroupVersionerIdentifier
//}
