package rest

import (
	"net/http"

	utilnet "k8s.io/apimachinery/pkg/util/net"
	"k8s.io/klog/v2"
)

// HTTPClientFor returns an http.Client that will provide the authentication
// or transport level security defined by the provided Config. Will return the
// default http.DefaultClient if no special case behavior is needed.
func HTTPClientFor(config *Config) (*http.Client, error) {
	transport, err := TransportFor(config)
	if err != nil {
		return nil, err
	}
	var httpClient *http.Client
	if transport != http.DefaultTransport || config.Timeout > 0 {
		httpClient = &http.Client{
			Transport: transport,
			Timeout:   config.Timeout,
		}
	} else {
		httpClient = http.DefaultClient
	}

	return httpClient, nil
}

// TransportFor returns an http.RoundTripper that will provide the authentication
// or transport level security defined by the provided Config. Will return the
// default http.DefaultTransport if no special case behavior is needed.
func TransportFor(config *Config) (http.RoundTripper, error) {
	var (
		rt http.RoundTripper
	)

	if config.Transport != nil {
		rt = config.Transport
	}

	if len(config.UserAgent) > 0 {
		rt = NewUserAgentRoundTripper(config.UserAgent, rt)
	}
	// TODO: 完善相关逻辑
	return rt, nil
}

// From K8s
type userAgentRoundTripper struct {
	agent string
	rt    http.RoundTripper
}

var _ utilnet.RoundTripperWrapper = &userAgentRoundTripper{}

// NewUserAgentRoundTripper will add User-Agent header to a request unless it has already been set.
func NewUserAgentRoundTripper(agent string, rt http.RoundTripper) http.RoundTripper {
	return &userAgentRoundTripper{agent, rt}
}

func (rt *userAgentRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	if len(req.Header.Get("User-Agent")) != 0 {
		return rt.rt.RoundTrip(req)
	}
	req = utilnet.CloneRequest(req)
	req.Header.Set("User-Agent", rt.agent)
	return rt.rt.RoundTrip(req)
}

func (rt *userAgentRoundTripper) CancelRequest(req *http.Request) {
	tryCancelRequest(rt.WrappedRoundTripper(), req)
}

func (rt *userAgentRoundTripper) WrappedRoundTripper() http.RoundTripper { return rt.rt }

func tryCancelRequest(rt http.RoundTripper, req *http.Request) {
	type canceler interface {
		CancelRequest(*http.Request)
	}
	switch rt := rt.(type) {
	case canceler:
		rt.CancelRequest(req)
	case utilnet.RoundTripperWrapper:
		tryCancelRequest(rt.WrappedRoundTripper(), req)
	default:
		klog.Warningf("Unable to cancel request for %T", rt)
	}
}
