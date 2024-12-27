package scheme

import (
	"hit.edu/framework/pkg/apimachinery/fields"
	"hit.edu/framework/pkg/apis/meta"
	"hit.edu/framework/pkg/apis/meta/internalversion"
	"net/url"
	"testing"
)

func TestConvert_v1_ListOptions_to_internalversion_ListOptions_fieldSelector(t *testing.T) {
	codec := ParameterCodec
	var tests = []struct {
		input    url.Values
		expected internalversion.ListOptions
	}{
		{
			input: url.Values{
				"fieldSelector": []string{"metadata.name=test"},
			},
			expected: internalversion.ListOptions{
				FieldSelector: fields.Set{"metadata.name": "test"}.AsSelector(),
			},
		},
	}
	from := meta.SchemeGroupVersion
	for _, test := range tests {
		opts := &internalversion.ListOptions{}
		err := codec.DecodeParameters(test.input, from, opts)
		if err != nil {
			t.Errorf("err:%s", err.Error())
		}
		if opts.FieldSelector.String() != test.expected.FieldSelector.String() {
			t.Errorf("expected:%#v, got:%#v", test.expected, opts)
		}
	}
}
