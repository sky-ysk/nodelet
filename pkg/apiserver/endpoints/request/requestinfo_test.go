package request

import (
	"k8s.io/apimachinery/pkg/util/sets"
	"net/http"
	"reflect"
	"testing"
)

func TestGetAPIRequestInfo(t *testing.T) {
	successCases := []struct {
		method              string
		url                 string
		expectedVerb        string
		expectedAPIPrefix   string
		expectedAPIGroup    string
		expectedAPIVersion  string
		expectedResource    string
		expectedSubresource string
		expectedName        string
		expectedNamespace   string
		expectedParts       []string
	}{

		{"GET", "/apis/resources/v1/nodes/foo", "get", "apis", "resources", "v1", "nodes", "", "foo", "", []string{"nodes", "foo"}},
		{"POST", "/apis/resources/v1/nodes", "create", "apis", "resources", "v1", "nodes", "", "", "", []string{"nodes"}},
		{"GET", "/apis/resources/v1/nodes", "list", "apis", "resources", "v1", "nodes", "", "", "", []string{"nodes"}},
		{"PUT", "/apis/resources/v1/nodes/foo", "update", "apis", "resources", "v1", "nodes", "", "foo", "", []string{"nodes", "foo"}},
		{"PATCH", "/apis/resources/v1/nodes/foo", "patch", "apis", "resources", "v1", "nodes", "", "foo", "", []string{"nodes", "foo"}},
		{"DELETE", "/apis/resources/v1/nodes/foo", "delete", "apis", "resources", "v1", "nodes", "", "foo", "", []string{"nodes", "foo"}},
		{"GET", "/apis/resources/v1/nodes/foo/status", "get", "apis", "resources", "v1", "nodes", "status", "foo", "", []string{"nodes", "foo", "status"}},
		{"GET", "/apis/resources/v1/namespaces/other/nodes/foo", "get", "apis", "resources", "v1", "nodes", "", "foo", "other", []string{"nodes", "foo"}},
		{"GET", "/apis/resources/v1/namespaces/other/nodes", "list", "apis", "resources", "v1", "nodes", "", "", "other", []string{"nodes"}},
		{"POST", "/apis/resources/v1/namespaces/other/nodes", "create", "apis", "resources", "v1", "nodes", "", "", "other", []string{"nodes"}},
		{"PUT", "/apis/resources/v1/namespaces/other/nodes/foo", "update", "apis", "resources", "v1", "nodes", "", "foo", "other", []string{"nodes", "foo"}},
		{"PATCH", "/apis/resources/v1/namespaces/other/nodes/foo", "patch", "apis", "resources", "v1", "nodes", "", "foo", "other", []string{"nodes", "foo"}},
		{"DELETE", "/apis/resources/v1/namespaces/other/nodes/foo", "delete", "apis", "resources", "v1", "nodes", "", "foo", "other", []string{"nodes", "foo"}},
		{"GET", "/apis/resources/v1/namespaces/other/nodes/foo/status", "get", "apis", "resources", "v1", "nodes", "status", "foo", "other", []string{"nodes", "foo", "status"}},
	}

	resolver := newTestRequestInfoResolver()

	for _, successCase := range successCases {
		req, _ := http.NewRequest(successCase.method, successCase.url, nil)

		apiRequestInfo, err := resolver.NewRequestInfo(req)
		if err != nil {
			t.Errorf("Unexpected error for url: %s %v", successCase.url, err)
		}
		if !apiRequestInfo.IsResourceRequest {
			t.Errorf("Expected resource request")
		}
		if successCase.expectedVerb != apiRequestInfo.Verb {
			t.Errorf("Unexpected verb for url: %s, expected: %s, actual: %s", successCase.url, successCase.expectedVerb, apiRequestInfo.Verb)
		}
		if successCase.expectedAPIVersion != apiRequestInfo.APIVersion {
			t.Errorf("Unexpected apiVersion for url: %s, expected: %s, actual: %s", successCase.url, successCase.expectedAPIVersion, apiRequestInfo.APIVersion)
		}
		if successCase.expectedResource != apiRequestInfo.Resource {
			t.Errorf("Unexpected resource for url: %s, expected: %s, actual: %s", successCase.url, successCase.expectedResource, apiRequestInfo.Resource)
		}
		if successCase.expectedSubresource != apiRequestInfo.Subresource {
			t.Errorf("Unexpected resource for url: %s, expected: %s, actual: %s", successCase.url, successCase.expectedSubresource, apiRequestInfo.Subresource)
		}
		if successCase.expectedName != apiRequestInfo.Name {
			t.Errorf("Unexpected name for url: %s, expected: %s, actual: %s", successCase.url, successCase.expectedName, apiRequestInfo.Name)
		}
		if successCase.expectedNamespace != apiRequestInfo.Namespace {
			t.Errorf("Unexpected namespace for url: %s, expected: %s, actual: %s", successCase.url, successCase.expectedNamespace, apiRequestInfo.Namespace)
		}
		if !reflect.DeepEqual(successCase.expectedParts, apiRequestInfo.Parts) {
			t.Errorf("Unexpected parts for url: %s, expected: %v, actual: %v", successCase.url, successCase.expectedParts, apiRequestInfo.Parts)
		}
	}

	errorCases := map[string]string{
		"no resource path":            "/",
		"just apiversion":             "/apis/version/",
		"just prefix, group, version": "/apis/group/version/",
		"bad prefix":                  "/badprefix/version/resource",
	}
	for k, v := range errorCases {
		req, err := http.NewRequest("GET", v, nil)
		if err != nil {
			t.Errorf("Unexpected error %v", err)
		}
		apiRequestInfo, err := resolver.NewRequestInfo(req)
		if err != nil {
			t.Errorf("%s: Unexpected error %v", k, err)
		}
		if apiRequestInfo.IsResourceRequest {
			t.Errorf("%s: expected non-resource request", k)
		}
	}
}

func newTestRequestInfoResolver() *RequestInfoFactory {
	return &RequestInfoFactory{
		APIPrefixes: sets.NewString("apis"),
	}
}
