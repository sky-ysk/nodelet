package meta

import (
	"hit.edu/framework/pkg/apimachinery/conversion"
	"hit.edu/framework/pkg/apimachinery/runtime"
	"net/url"
	"unsafe"
)

func init() {
	localSchemeBuilder.Register(RegisterConversions)
}

func RegisterConversions(s *runtime.Scheme) error {
	//TODO:实现fieldSelector、labelSelector字符串到结构体的转换
	if err := s.AddGeneratedConversionFunc((*url.Values)(nil), (*CreateOptions)(nil), func(a, b interface{}, scope conversion.Scope) error {
		return Convert_url_Values_To_v1_CreateOptions(a.(*url.Values), b.(*CreateOptions), scope)
	}); err != nil {
		return err
	}
	if err := s.AddGeneratedConversionFunc((*url.Values)(nil), (*DeleteOptions)(nil), func(a, b interface{}, scope conversion.Scope) error {
		return Convert_url_Values_To_v1_DeleteOptions(a.(*url.Values), b.(*DeleteOptions), scope)
	}); err != nil {
		return err
	}
	if err := s.AddGeneratedConversionFunc((*url.Values)(nil), (*GetOptions)(nil), func(a, b interface{}, scope conversion.Scope) error {
		return Convert_url_Values_To_v1_GetOptions(a.(*url.Values), b.(*GetOptions), scope)
	}); err != nil {
		return err
	}
	if err := s.AddGeneratedConversionFunc((*url.Values)(nil), (*ListOptions)(nil), func(a, b interface{}, scope conversion.Scope) error {
		return Convert_url_Values_To_v1_ListOptions(a.(*url.Values), b.(*ListOptions), scope)
	}); err != nil {
		return err
	}
	if err := s.AddGeneratedConversionFunc((*url.Values)(nil), (*UpdateOptions)(nil), func(a, b interface{}, scope conversion.Scope) error {
		return Convert_url_Values_To_v1_UpdateOptions(a.(*url.Values), b.(*UpdateOptions), scope)
	}); err != nil {
		return err
	}
	if err := s.AddGeneratedConversionFunc((*url.Values)(nil), (*PatchOptions)(nil), func(a, b interface{}, scope conversion.Scope) error {
		return Convert_url_Values_To_v1_PatchOptions(a.(*url.Values), b.(*PatchOptions), scope)
	}); err != nil {
		return err
	}
	
	return nil
}

func Convert_url_Values_To_v1_CreateOptions(in *url.Values, out *CreateOptions, s conversion.Scope) error {
	if values, ok := map[string][]string(*in)["dryRun"]; ok && len(values) > 0 {
		out.DryRun = *(*[]string)(unsafe.Pointer(&values))
	} else {
		out.DryRun = nil
	}
	if values, ok := map[string][]string(*in)["fieldManager"]; ok && len(values) > 0 {
		if err := runtime.Convert_Slice_string_To_string(&values, &out.FieldManager, s); err != nil {
			return err
		}
	} else {
		out.FieldManager = ""
	}
	if values, ok := map[string][]string(*in)["fieldValidation"]; ok && len(values) > 0 {
		if err := runtime.Convert_Slice_string_To_string(&values, &out.FieldValidation, s); err != nil {
			return err
		}
	} else {
		out.FieldValidation = ""
	}
	return nil
}

func Convert_url_Values_To_v1_DeleteOptions(in *url.Values, out *DeleteOptions, s conversion.Scope) error {
	if err := autoConvert_url_Values_To_v1_DeleteOptions(in, out, s); err != nil {
		return err
	}
	
	uid := UID("")
	if values, ok := (*in)["uid"]; ok && len(values) > 0 {
		uid = UID(values[0])
	}
	
	resourceVersion := ""
	if values, ok := (*in)["resourceVersion"]; ok && len(values) > 0 {
		resourceVersion = values[0]
	}
	
	if len(uid) > 0 || len(resourceVersion) > 0 {
		if out.Preconditions == nil {
			out.Preconditions = &Preconditions{}
		}
		if len(uid) > 0 {
			out.Preconditions.UID = &uid
		}
		if len(resourceVersion) > 0 {
			out.Preconditions.ResourceVersion = &resourceVersion
		}
	}
	return nil
}

func autoConvert_url_Values_To_v1_DeleteOptions(in *url.Values, out *DeleteOptions, s conversion.Scope) error {
	// WARNING: Field TypeMeta does not have json tag, skipping.
	
	if values, ok := map[string][]string(*in)["gracePeriodSeconds"]; ok && len(values) > 0 {
		if err := runtime.Convert_Slice_string_To_Pointer_int64(&values, &out.GracePeriodSeconds, s); err != nil {
			return err
		}
	} else {
		out.GracePeriodSeconds = nil
	}
	// INFO: in.Preconditions opted out of conversion generation
	if values, ok := map[string][]string(*in)["orphanDependents"]; ok && len(values) > 0 {
		if err := runtime.Convert_Slice_string_To_Pointer_bool(&values, &out.OrphanDependents, s); err != nil {
			return err
		}
	} else {
		out.OrphanDependents = nil
	}
	if values, ok := map[string][]string(*in)["propagationPolicy"]; ok && len(values) > 0 {
		if err := Convert_Slice_string_To_Pointer_v1_DeletionPropagation(&values, &out.PropagationPolicy, s); err != nil {
			return err
		}
	} else {
		out.PropagationPolicy = nil
	}
	if values, ok := map[string][]string(*in)["dryRun"]; ok && len(values) > 0 {
		out.DryRun = *(*[]string)(unsafe.Pointer(&values))
	} else {
		out.DryRun = nil
	}
	return nil
}

// Convert_Slice_string_To_Pointer_v1_DeletionPropagation allows converting a URL query parameter propagationPolicy
func Convert_Slice_string_To_Pointer_v1_DeletionPropagation(in *[]string, out **DeletionPropagation, s conversion.Scope) error {
	var str string
	if len(*in) > 0 {
		str = (*in)[0]
	} else {
		str = ""
	}
	temp := DeletionPropagation(str)
	*out = &temp
	return nil
}

// Convert_url_Values_To_v1_GetOptions is an autogenerated conversion function.
func Convert_url_Values_To_v1_GetOptions(in *url.Values, out *GetOptions, s conversion.Scope) error {
	if values, ok := map[string][]string(*in)["resourceVersion"]; ok && len(values) > 0 {
		if err := runtime.Convert_Slice_string_To_string(&values, &out.ResourceVersion, s); err != nil {
			return err
		}
	} else {
		out.ResourceVersion = ""
	}
	return nil
}

func Convert_url_Values_To_v1_UpdateOptions(in *url.Values, out *UpdateOptions, s conversion.Scope) error {
	if values, ok := map[string][]string(*in)["dryRun"]; ok && len(values) > 0 {
		out.DryRun = *(*[]string)(unsafe.Pointer(&values))
	} else {
		out.DryRun = nil
	}
	if values, ok := map[string][]string(*in)["fieldManager"]; ok && len(values) > 0 {
		if err := runtime.Convert_Slice_string_To_string(&values, &out.FieldManager, s); err != nil {
			return err
		}
	} else {
		out.FieldManager = ""
	}
	if values, ok := map[string][]string(*in)["fieldValidation"]; ok && len(values) > 0 {
		if err := runtime.Convert_Slice_string_To_string(&values, &out.FieldValidation, s); err != nil {
			return err
		}
	} else {
		out.FieldValidation = ""
	}
	return nil
}

func Convert_url_Values_To_v1_ListOptions(in *url.Values, out *ListOptions, s conversion.Scope) error {
	if values, ok := map[string][]string(*in)["labelSelector"]; ok && len(values) > 0 {
		if err := runtime.Convert_Slice_string_To_string(&values, &out.LabelSelector, s); err != nil {
			return err
		}
	} else {
		out.LabelSelector = ""
	}
	if values, ok := map[string][]string(*in)["fieldSelector"]; ok && len(values) > 0 {
		if err := runtime.Convert_Slice_string_To_string(&values, &out.FieldSelector, s); err != nil {
			return err
		}
	} else {
		out.FieldSelector = ""
	}
	if values, ok := map[string][]string(*in)["watch"]; ok && len(values) > 0 {
		if err := runtime.Convert_Slice_string_To_bool(&values, &out.Watch, s); err != nil {
			return err
		}
	} else {
		out.Watch = false
	}
	if values, ok := map[string][]string(*in)["allowWatchBookmarks"]; ok && len(values) > 0 {
		if err := runtime.Convert_Slice_string_To_bool(&values, &out.AllowWatchBookmarks, s); err != nil {
			return err
		}
	} else {
		out.AllowWatchBookmarks = false
	}
	if values, ok := map[string][]string(*in)["progressNotify"]; ok && len(values) > 0 {
		if err := runtime.Convert_Slice_string_To_bool(&values, &out.ProgressNotify, s); err != nil {
			return err
		}
	} else {
		out.ProgressNotify = false
	}
	if values, ok := map[string][]string(*in)["resourceVersion"]; ok && len(values) > 0 {
		if err := runtime.Convert_Slice_string_To_string(&values, &out.ResourceVersion, s); err != nil {
			return err
		}
	} else {
		out.ResourceVersion = ""
	}
	if values, ok := map[string][]string(*in)["resourceVersionMatch"]; ok && len(values) > 0 {
		if err := Convert_Slice_string_To_v1_ResourceVersionMatch(&values, &out.ResourceVersionMatch, s); err != nil {
			return err
		}
	} else {
		out.ResourceVersionMatch = ""
	}
	if values, ok := map[string][]string(*in)["timeoutSeconds"]; ok && len(values) > 0 {
		if err := runtime.Convert_Slice_string_To_Pointer_int64(&values, &out.TimeoutSeconds, s); err != nil {
			return err
		}
	} else {
		out.TimeoutSeconds = nil
	}
	if values, ok := map[string][]string(*in)["limit"]; ok && len(values) > 0 {
		if err := runtime.Convert_Slice_string_To_int64(&values, &out.Limit, s); err != nil {
			return err
		}
	} else {
		out.Limit = 0
	}
	if values, ok := map[string][]string(*in)["continue"]; ok && len(values) > 0 {
		if err := runtime.Convert_Slice_string_To_string(&values, &out.Continue, s); err != nil {
			return err
		}
	} else {
		out.Continue = ""
	}
	if values, ok := map[string][]string(*in)["sendInitialEvents"]; ok && len(values) > 0 {
		if err := runtime.Convert_Slice_string_To_Pointer_bool(&values, &out.SendInitialEvents, s); err != nil {
			return err
		}
	} else {
		out.SendInitialEvents = nil
	}
	return nil
}

// Convert_url_Values_To_v1_PatchOptions is an autogenerated conversion function.
func Convert_url_Values_To_v1_PatchOptions(in *url.Values, out *PatchOptions, s conversion.Scope) error {
	if values, ok := map[string][]string(*in)["dryRun"]; ok && len(values) > 0 {
		out.DryRun = *(*[]string)(unsafe.Pointer(&values))
	} else {
		out.DryRun = nil
	}
	if values, ok := map[string][]string(*in)["fieldValidation"]; ok && len(values) > 0 {
		if err := runtime.Convert_Slice_string_To_string(&values, &out.FieldValidation, s); err != nil {
			return err
		}
	} else {
		out.FieldValidation = ""
	}
	return nil
}

func Convert_Slice_string_To_v1_ResourceVersionMatch(in *[]string, out *ResourceVersionMatch, s conversion.Scope) error {
	if len(*in) > 0 {
		*out = ResourceVersionMatch((*in)[0])
	}
	return nil
}
