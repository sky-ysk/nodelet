//package text
//
//import (
//	"fmt"
//	"hit.edu/framework/pkg/apimachinery/runtime"
//	"io"
//)
//
//// PlainTextSerializer 用于处理 text/plain 类型
//type Serializer struct{}
//
//var _ runtime.Serializer = &Serializer{}
//
//// Serializer 用于处理 text/plain 类型的序列化和反序列化
//// Encode 将字节数据（String 类型）转换为原始的 byte 数据
//func (s *Serializer) Encode(obj runtime.Object, writer io.Writer) error {
//	// 检查 obj 是否是我们期望的类型：*PlainTextObject
//	if plainTextObj, ok := obj.(*PlainTextObject); ok {
//		// 将 plainTextObj.Data 转换为字节并写入到 writer 中
//		_, err := writer.Write([]byte(plainTextObj.Data))
//		return err
//	}
//	// 如果类型不匹配，返回错误
//	return fmt.Errorf("Serializer only supports *PlainTextObject type, but got %T", obj)
//}
//
//// Decode 将字节数据转换为 PlainTextObject 类型
//func (s *Serializer) Decode(data []byte, into runtime.Object) (runtime.Object, error) {
//	// 检查 into 是否是 *PlainTextObject 类型
//	if plainTextObj, ok := into.(*PlainTextObject); ok {
//		// 将字节数据转换为字符串并赋值给 PlainTextObject 的 Data 字段
//		plainTextObj.Data = string(data)
//		return plainTextObj, nil
//	}
//	// 如果 into 类型不匹配，返回错误
//	return nil, fmt.Errorf("Serializer only supports *PlainTextObject type for decoding, but got %T", into)
//}
//
//// PlainTextObject 用于存储字符串数据，适配 runtime.Object 接口
//type PlainTextObject struct {
//	Data string
//}
