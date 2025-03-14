package runtime

import (
	"io"
	"net/url"

	"hit.edu/framework/pkg/apimachinery/runtime/schema"
)

const (
	// APIVersionInternal may be used if you are registering a type that should not
	// be considered stable or serialized - it is a convention only and has no
	// special behavior in this package.
	APIVersionInternal = "__internal"
)

type Object interface {
	GetObjectKind() schema.ObjectKind
	DeepCopyObject() Object
}

// Namer provides methods for retrieving name and namespace of an API object.
type Namer interface {
	// Name returns the name of a given object.
	Name(obj Object) (string, error)
	// Namespace returns the name of a given object.
	Namespace(obj Object) (string, error)
}

type Serializer interface {
	Encoder
	Decoder
}

type Encoder interface {
	Encode(obj Object, writer io.Writer) error

	// Identifier returns an identifier of the encoder.
	// Identifiers of two different encoders should be equal if and only if for every input
	// object it will be encoded to the same representation by both of them.
	//
	// Identifier is intended for use with CacheableObject#CacheEncode method. In order to
	// correctly handle CacheableObject, Encode() method should look similar to below, where
	// doEncode() is the encoding logic of implemented encoder:
	//   func (e *MyEncoder) Encode(obj Object, w io.Writer) error {
	//     if co, ok := obj.(CacheableObject); ok {
	//       return co.CacheEncode(e.Identifier(), e.doEncode, w)
	//     }
	//     return e.doEncode(obj, w)
	//   }
	Identifier() Identifier
}

type Decoder interface {
	Decode(data []byte, defaults *schema.GroupVersionKind, into Object) (Object, *schema.GroupVersionKind, error)
}

type Codec Serializer

type ParameterCodec interface {
	// DecodeParameters takes the given url.Values in the specified group version and decodes them
	// into the provided object, or returns an error.
	DecodeParameters(parameters url.Values, from schema.GroupVersion, into Object) error
	// EncodeParameters encodes the provided object as query parameters or returns an error.
	EncodeParameters(obj Object, to schema.GroupVersion) (url.Values, error)
}

// ClientNegotiator handles turning an HTTP content type into the appropriate encoder.
// Use NewClientNegotiator or NewVersionedClientNegotiator to create this interface from
// a NegotiatedSerializer.
// ClientNegotiator 在request的body()中用到
type ClientNegotiator interface {
	// Encoder returns the appropriate encoder for the provided contentType (e.g. application/json)
	// and any optional mediaType parameters (e.g. pretty=1), or an error. If no serializer is found
	// a NegotiateError will be returned. The current client implementations consider params to be
	// optional modifiers to the contentType and will ignore unrecognized parameters.
	Encoder(contentType string, params map[string]string) (Encoder, error)
	// Decoder returns the appropriate decoder for the provided contentType (e.g. application/json)
	// and any optional mediaType parameters (e.g. pretty=1), or an error. If no serializer is found
	// a NegotiateError will be returned. The current client implementations consider params to be
	// optional modifiers to the contentType and will ignore unrecognized parameters.
	Decoder(contentType string, params map[string]string) (Decoder, error)
	// StreamDecoder returns the appropriate stream decoder for the provided contentType (e.g.
	// application/json) and any optional mediaType parameters (e.g. pretty=1), or an error. If no
	// serializer is found a NegotiateError will be returned. The Serializer and Framer will always
	// be returned if a Decoder is returned. The current client implementations consider params to be
	// optional modifiers to the contentType and will ignore unrecognized parameters.
	StreamDecoder(contentType string, params map[string]string) (Decoder, Serializer, Framer, error)
}

// Framer is a factory for creating readers and writers that obey a particular framing pattern.
// 在上面的ClientNegotiator 中用到
type Framer interface {
	NewFrameReader(r io.ReadCloser) io.ReadCloser
	NewFrameWriter(w io.Writer) io.Writer
}

// 在negotiate.go中用到
// NegotiatedSerializer is an interface used for obtaining encoders, decoders, and serializers
// for multiple supported media types. This would commonly be accepted by a server component
// that performs HTTP content negotiation to accept multiple formats.
type NegotiatedSerializer interface {
	// SupportedMediaTypes is the media types supported for reading and writing single objects.
	SupportedMediaTypes() []SerializerInfo

	// EncoderForVersion returns an encoder that ensures objects being written to the provided
	// serializer are in the provided group version.
	EncoderForVersion(serializer Encoder, gv GroupVersioner) Encoder
	// DecoderToVersion returns a decoder that ensures objects being read by the provided
	// serializer are in the provided group version by default.
	DecoderToVersion(serializer Decoder, gv GroupVersioner) Decoder
}

// 在NegotiatedSerializer 中用到
// SerializerInfo contains information about a specific serialization format
type SerializerInfo struct {
	// MediaType is the value that represents this serializer over the wire.
	MediaType string
	// MediaTypeType is the first part of the MediaType ("application" in "application/json").
	MediaTypeType string
	// MediaTypeSubType is the second part of the MediaType ("json" in "application/json").
	MediaTypeSubType string
	// EncodesAsText indicates this serializer can be encoded to UTF-8 safely.
	EncodesAsText bool
	// Serializer is the individual object serializer for this media type.
	Serializer Serializer
	// PrettySerializer, if set, can serialize this object in a form biased towards
	// readability.
	PrettySerializer Serializer
	// StrictSerializer, if set, deserializes this object strictly,
	// erring on unknown fields.
	StrictSerializer Serializer
	// StreamSerializer, if set, describes the streaming serialization format
	// for this media type.
	StreamSerializer *StreamSerializerInfo
}

// 在SerializerInfo中用到
// StreamSerializerInfo contains information about a specific stream serialization format
type StreamSerializerInfo struct {
	// EncodesAsText indicates this serializer can be encoded to UTF-8 safely.
	EncodesAsText bool
	// Serializer is the top level object serializer for this type when streaming
	Serializer
	// Framer is the factory for retrieving streams that separate objects on the wire
	Framer
}

// 在NegotiatedSerializer中引用
// GroupVersioner refines a set of possible conversion targets into a single option.
type GroupVersioner interface {
	// KindForGroupVersionKinds returns a desired target group version kind for the given input, or returns ok false if no
	// target is known. In general, if the return target is not in the input list, the caller is expected to invoke
	// Scheme.New(target) and then perform a conversion between the current Go type and the destination Go type.
	// Sophisticated implementations may use additional information about the input kinds to pick a destination kind.
	KindForGroupVersionKinds(kinds []schema.GroupVersionKind) (target schema.GroupVersionKind, ok bool)
	// Identifier returns string representation of the object.
	// Identifiers of two different encoders should be equal only if for every input
	// kinds they return the same result.
	Identifier() string
}

// ObjectCreater contains methods for instantiating an object by kind and version.
type ObjectCreater interface {
	New(kind schema.GroupVersionKind) (out Object, err error)
}

// ObjectTyper contains methods for extracting the APIVersion and Kind
// of objects.
type ObjectTyper interface {
	// ObjectKinds returns the all possible group,version,kind of the provided object, true if
	// the object is unversioned, or an error if the object is not recognized
	// (IsNotRegisteredError will return true).
	ObjectKinds(Object) ([]schema.GroupVersionKind, bool, error)
	// Recognizes returns true if the scheme is able to handle the provided version and kind,
	// or more precisely that the provided version is a possible conversion or decoding
	// target.
	Recognizes(gvk schema.GroupVersionKind) bool
}

// Identifier represents an identifier.
// Identitier of two different objects should be equal if and only if for every
// input the output they produce is exactly the same.
type Identifier string

// Unstructured objects store values as map[string]interface{}, with only values that can be serialized
// to JSON allowed.
type Unstructured interface {
	Object
	// NewEmptyInstance returns a new instance of the concrete type containing only kind/apiVersion and no other data.
	// This should be called instead of reflect.New() for unstructured types because the go type alone does not preserve kind/apiVersion info.
	NewEmptyInstance() Unstructured
	// UnstructuredContent returns a non-nil map with this object's contents. Values may be
	// []interface{}, map[string]interface{}, or any primitive type. Contents are typically serialized to
	// and from JSON. SetUnstructuredContent should be used to mutate the contents.
	UnstructuredContent() map[string]interface{}
	// SetUnstructuredContent updates the object content to match the provided map.
	SetUnstructuredContent(map[string]interface{})
	// IsList returns true if this type is a list or matches the list convention - has an array called "items".
	IsList() bool
	// EachListItem should pass a single item out of the list as an Object to the provided function. Any
	// error should terminate the iteration. If IsList() returns false, this method should return an error
	// instead of calling the provided function.
	EachListItem(func(Object) error) error
	// EachListItemWithAlloc works like EachListItem, but avoids retaining references to a slice of items.
	// It does this by making a shallow copy of non-pointer items before passing them to fn.
	//
	// If the items passed to fn are not retained, or are retained for the same duration, use EachListItem instead for memory efficiency.
	EachListItemWithAlloc(func(Object) error) error
}

// CacheableObject allows an object to cache its different serializations
// to avoid performing the same serialization multiple times.
type CacheableObject interface {
	// CacheEncode writes an object to a stream. The <encode> function will
	// be used in case of cache miss. The <encode> function takes ownership
	// of the object.
	// If CacheableObject is a wrapper, then deep-copy of the wrapped object
	// should be passed to <encode> function.
	// CacheEncode assumes that for two different calls with the same <id>,
	// <encode> function will also be the same.
	CacheEncode(id Identifier, encode func(Object, io.Writer) error, w io.Writer) error
	// GetObject returns a deep-copy of an object to be encoded - the caller of
	// GetObject() is the owner of returned object. The reason for making a copy
	// is to avoid bugs, where caller modifies the object and forgets to copy it,
	// thus modifying the object for everyone.
	// The object returned by GetObject should be the same as the one that is supposed
	// to be passed to <encode> function in CacheEncode method.
	// If CacheableObject is a wrapper, the copy of wrapped object should be returned.
	GetObject() Object
}

// ObjectConvertor converts an object to a different version.
type ObjectConvertor interface {
	// Convert attempts to convert one object into another, or returns an error. This
	// method does not mutate the in object, but the in and out object might share data structures,
	// i.e. the out object cannot be mutated without mutating the in object as well.
	// The context argument will be passed to all nested conversions.
	Convert(in, out, context interface{}) error
	// ConvertToVersion takes the provided object and converts it the provided version. This
	// method does not mutate the in object, but the in and out object might share data structures,
	// i.e. the out object cannot be mutated without mutating the in object as well.
	// This method is similar to Convert() but handles specific details of choosing the correct
	// output version.

	ConvertToVersion(in Object, gv GroupVersioner) (out Object, err error)
	ConvertFieldLabel(gvk schema.GroupVersionKind, label, value string) (string, string, error)
}

type ObjectDefaulter interface {
	// Default takes an object (must be a pointer) and applies any default values.
	// Defaulters may not error.
	Default(in Object)
}

// MemoryAllocator is responsible for allocating memory.
// By encapsulating memory allocation into its own interface, we can reuse the memory
// across many operations in places we know it can significantly improve the performance.
type MemoryAllocator interface {
	// Allocate reserves memory for n bytes.
	// Note that implementations of this method are not required to zero the returned array.
	// It is the caller's responsibility to clean the memory if needed.
	Allocate(n uint64) []byte
}

// EncoderWithAllocator  serializes objects in a way that allows callers to manage any additional memory allocations.
type EncoderWithAllocator interface {
	Encoder
	// EncodeWithAllocator writes an object to a stream as Encode does.
	// In addition, it allows for providing a memory allocator for efficient memory usage during object serialization
	EncodeWithAllocator(obj Object, w io.Writer, memAlloc MemoryAllocator) error
}

// NestedObjectEncoder is an optional interface that objects may implement to be given
// an opportunity to encode any nested Objects / RawExtensions during serialization.
type NestedObjectEncoder interface {
	EncodeNestedObjects(e Encoder) error
}

// NestedObjectDecoder is an optional interface that objects may implement to be given
// an opportunity to decode any nested Objects / RawExtensions during serialization.
// It is possible for DecodeNestedObjects to return a non-nil error but for the decoding
// to have succeeded in the case of strict decoding errors (e.g. unknown/duplicate fields).
// As such it is important for callers of DecodeNestedObjects to check to confirm whether
// an error is a runtime.StrictDecodingError before short circuiting.
// Similarly, implementations of DecodeNestedObjects should ensure that a runtime.StrictDecodingError
// is only returned when the rest of decoding has succeeded.
type NestedObjectDecoder interface {
	DecodeNestedObjects(d Decoder) error
}
