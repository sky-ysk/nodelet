/*
Copyright 2014 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// Package runtime defines conversions between generic types and structs to map query strings
// to struct objects.
package meta

import (
	"fmt"
	"hit.edu/framework/pkg/apimachinery/conversion"
	"hit.edu/framework/pkg/apimachinery/fields"
	"hit.edu/framework/pkg/apimachinery/labels"
	"hit.edu/framework/pkg/apimachinery/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
	"reflect"
	"strconv"
	"strings"
)

// DefaultMetaV1FieldSelectorConversion auto-accepts metav1 values for name and namespace.
// A cluster scoped resource specifying namespace empty works fine and specifying a particular
// namespace will return no results, as expected.
func DefaultMetaV1FieldSelectorConversion(label, value string) (string, string, error) {
	switch label {
	case "metadata.name":
		return label, value, nil
	case "metadata.namespace":
		return label, value, nil
	default:
		return "", "", fmt.Errorf("%q is not a known field selector: only %q, %q", label, "metadata.name", "metadata.namespace")
	}
}

// JSONKeyMapper uses the struct tags on a conversion to determine the key value for
// the other side. Use when mapping from a map[string]* to a struct or vice versa.
func JSONKeyMapper(key string, sourceTag, destTag reflect.StructTag) (string, string) {
	if s := destTag.Get("json"); len(s) > 0 {
		return strings.SplitN(s, ",", 2)[0], key
	}
	if s := sourceTag.Get("json"); len(s) > 0 {
		return key, strings.SplitN(s, ",", 2)[0]
	}
	return key, key
}

func Convert_Slice_string_To_string(in *[]string, out *string, s conversion.Scope) error {
	if len(*in) == 0 {
		*out = ""
		return nil
	}
	*out = (*in)[0]
	return nil
}

func Convert_Slice_string_To_int(in *[]string, out *int, s conversion.Scope) error {
	if len(*in) == 0 {
		*out = 0
		return nil
	}
	str := (*in)[0]
	i, err := strconv.Atoi(str)
	if err != nil {
		return err
	}
	*out = i
	return nil
}

// Convert_Slice_string_To_bool will convert a string parameter to boolean.
// Only the absence of a value (i.e. zero-length slice), a value of "false", or a
// value of "0" resolve to false.
// Any other value (including empty string) resolves to true.
func Convert_Slice_string_To_bool(in *[]string, out *bool, s conversion.Scope) error {
	if len(*in) == 0 {
		*out = false
		return nil
	}
	switch {
	case (*in)[0] == "0", strings.EqualFold((*in)[0], "false"):
		*out = false
	default:
		*out = true
	}
	return nil
}

// Convert_Slice_string_To_bool will convert a string parameter to boolean.
// Only the absence of a value (i.e. zero-length slice), a value of "false", or a
// value of "0" resolve to false.
// Any other value (including empty string) resolves to true.
func Convert_Slice_string_To_Pointer_bool(in *[]string, out **bool, s conversion.Scope) error {
	if len(*in) == 0 {
		boolVar := false
		*out = &boolVar
		return nil
	}
	switch {
	case (*in)[0] == "0", strings.EqualFold((*in)[0], "false"):
		boolVar := false
		*out = &boolVar
	default:
		boolVar := true
		*out = &boolVar
	}
	return nil
}

func string_to_int64(in string) (int64, error) {
	return strconv.ParseInt(in, 10, 64)
}

func Convert_string_To_int64(in *string, out *int64, s conversion.Scope) error {
	if in == nil {
		*out = 0
		return nil
	}
	i, err := string_to_int64(*in)
	if err != nil {
		return err
	}
	*out = i
	return nil
}

func Convert_Slice_string_To_int64(in *[]string, out *int64, s conversion.Scope) error {
	if len(*in) == 0 {
		*out = 0
		return nil
	}
	i, err := string_to_int64((*in)[0])
	if err != nil {
		return err
	}
	*out = i
	return nil
}

func Convert_string_To_Pointer_int64(in *string, out **int64, s conversion.Scope) error {
	if in == nil {
		*out = nil
		return nil
	}
	i, err := string_to_int64(*in)
	if err != nil {
		return err
	}
	*out = &i
	return nil
}

func Convert_Slice_string_To_Pointer_int64(in *[]string, out **int64, s conversion.Scope) error {
	if len(*in) == 0 {
		*out = nil
		return nil
	}
	i, err := string_to_int64((*in)[0])
	if err != nil {
		return err
	}
	*out = &i
	return nil
}

func RegisterStringConversions(s *runtime.Scheme) error {
	if err := s.AddConversionFunc((*[]string)(nil), (*string)(nil), func(a, b interface{}, scope conversion.Scope) error {
		return Convert_Slice_string_To_string(a.(*[]string), b.(*string), scope)
	}); err != nil {
		return err
	}
	if err := s.AddConversionFunc((*[]string)(nil), (*int)(nil), func(a, b interface{}, scope conversion.Scope) error {
		return Convert_Slice_string_To_int(a.(*[]string), b.(*int), scope)
	}); err != nil {
		return err
	}
	if err := s.AddConversionFunc((*[]string)(nil), (*bool)(nil), func(a, b interface{}, scope conversion.Scope) error {
		return Convert_Slice_string_To_bool(a.(*[]string), b.(*bool), scope)
	}); err != nil {
		return err
	}
	if err := s.AddConversionFunc((*[]string)(nil), (*int64)(nil), func(a, b interface{}, scope conversion.Scope) error {
		return Convert_Slice_string_To_int64(a.(*[]string), b.(*int64), scope)
	}); err != nil {
		return err
	}
	return nil
}

func Convert_Pointer_float64_To_float64(in **float64, out *float64, s conversion.Scope) error {
	if *in == nil {
		*out = 0
		return nil
	}
	*out = float64(**in)
	return nil
}

func Convert_float64_To_Pointer_float64(in *float64, out **float64, s conversion.Scope) error {
	temp := float64(*in)
	*out = &temp
	return nil
}

func Convert_Pointer_int32_To_int32(in **int32, out *int32, s conversion.Scope) error {
	if *in == nil {
		*out = 0
		return nil
	}
	*out = int32(**in)
	return nil
}

func Convert_int32_To_Pointer_int32(in *int32, out **int32, s conversion.Scope) error {
	temp := int32(*in)
	*out = &temp
	return nil
}

func Convert_Pointer_int64_To_int64(in **int64, out *int64, s conversion.Scope) error {
	if *in == nil {
		*out = 0
		return nil
	}
	*out = int64(**in)
	return nil
}

func Convert_int64_To_Pointer_int64(in *int64, out **int64, s conversion.Scope) error {
	temp := int64(*in)
	*out = &temp
	return nil
}

func Convert_Pointer_int64_To_int(in **int64, out *int, s conversion.Scope) error {
	if *in == nil {
		*out = 0
		return nil
	}
	*out = int(**in)
	return nil
}

func Convert_int_To_Pointer_int64(in *int, out **int64, s conversion.Scope) error {
	temp := int64(*in)
	*out = &temp
	return nil
}

func Convert_Pointer_string_To_string(in **string, out *string, s conversion.Scope) error {
	if *in == nil {
		*out = ""
		return nil
	}
	*out = **in
	return nil
}

func Convert_string_To_Pointer_string(in *string, out **string, s conversion.Scope) error {
	if in == nil {
		stringVar := ""
		*out = &stringVar
		return nil
	}
	*out = in
	return nil
}

func Convert_Pointer_bool_To_bool(in **bool, out *bool, s conversion.Scope) error {
	if *in == nil {
		*out = false
		return nil
	}
	*out = **in
	return nil
}

func Convert_bool_To_Pointer_bool(in *bool, out **bool, s conversion.Scope) error {
	if in == nil {
		boolVar := false
		*out = &boolVar
		return nil
	}
	*out = in
	return nil
}

// +k8s:conversion-fn=drop
func Convert_v1_TypeMeta_To_v1_TypeMeta(in, out *TypeMeta, s conversion.Scope) error {
	// These values are explicitly not copied
	//out.APIVersion = in.APIVersion
	//out.Kind = in.Kind
	return nil
}

// +k8s:conversion-fn=copy-only
func Convert_v1_ListMeta_To_v1_ListMeta(in, out *ListMeta, s conversion.Scope) error {
	*out = *in
	return nil
}

// +k8s:conversion-fn=copy-only
func Convert_v1_DeleteOptions_To_v1_DeleteOptions(in, out *DeleteOptions, s conversion.Scope) error {
	*out = *in
	return nil
}

// +k8s:conversion-fn=copy-only
func Convert_intstr_IntOrString_To_intstr_IntOrString(in, out *intstr.IntOrString, s conversion.Scope) error {
	*out = *in
	return nil
}

func Convert_Pointer_intstr_IntOrString_To_intstr_IntOrString(in **intstr.IntOrString, out *intstr.IntOrString, s conversion.Scope) error {
	if *in == nil {
		*out = intstr.IntOrString{} // zero value
		return nil
	}
	*out = **in // copy
	return nil
}

func Convert_intstr_IntOrString_To_Pointer_intstr_IntOrString(in *intstr.IntOrString, out **intstr.IntOrString, s conversion.Scope) error {
	temp := *in // copy
	*out = &temp
	return nil
}

// Convert_Slice_string_To_v1_Time allows converting a URL query parameter value
func Convert_Slice_string_To_v1_Time(in *[]string, out *Time, s conversion.Scope) error {
	str := ""
	if len(*in) > 0 {
		str = (*in)[0]
	}
	return out.UnmarshalQueryParameter(str)
}

func Convert_Slice_string_To_Pointer_v1_Time(in *[]string, out **Time, s conversion.Scope) error {
	if in == nil {
		return nil
	}
	str := ""
	if len(*in) > 0 {
		str = (*in)[0]
	}
	temp := Time{}
	if err := temp.UnmarshalQueryParameter(str); err != nil {
		return err
	}
	*out = &temp
	return nil
}

func Convert_string_To_labels_Selector(in *string, out *labels.Selector, s conversion.Scope) error {
	selector, err := labels.Parse(*in)
	if err != nil {
		return err
	}
	*out = selector
	return nil
}

func Convert_string_To_fields_Selector(in *string, out *fields.Selector, s conversion.Scope) error {
	selector, err := fields.ParseSelector(*in)
	if err != nil {
		return err
	}
	*out = selector
	return nil
}

func Convert_labels_Selector_To_string(in *labels.Selector, out *string, s conversion.Scope) error {
	if *in == nil {
		return nil
	}
	*out = (*in).String()
	return nil
}

func Convert_fields_Selector_To_string(in *fields.Selector, out *string, s conversion.Scope) error {
	if *in == nil {
		return nil
	}
	*out = (*in).String()
	return nil
}

// Convert_Slice_string_To_Slice_int32 converts multiple query parameters or
// a single query parameter with a comma delimited value to multiple int32.
// This is used for port forwarding which needs the ports as int32.
func Convert_Slice_string_To_Slice_int32(in *[]string, out *[]int32, s conversion.Scope) error {
	for _, s := range *in {
		for _, v := range strings.Split(s, ",") {
			x, err := strconv.ParseUint(v, 10, 16)
			if err != nil {
				return fmt.Errorf("cannot convert to []int32: %v", err)
			}
			*out = append(*out, int32(x))
		}
	}
	return nil
}
