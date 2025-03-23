// errors.go主要负责处理与store操作相关的部分，包括与资源版本和分页相关的错误
package etcd3

import (
	"encoding/json"
	"net/http"

	etcdrpc "go.etcd.io/etcd/api/v3/v3rpc/rpctypes"
	"hit.edu/framework/pkg/apiserver/registry/storage"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
)

func interpretWatchError(err error) error {
	switch {
	case err == etcdrpc.ErrCompacted:
		return errors.NewResourceExpired("The resourceVersion for the provided watch is too old.")
	}
	return err
}

const (
	expired         string = "The resourceVersion for the provided list is too old."
	continueExpired string = "The provided continue parameter is too old " +
		"to display a consistent list result. You can start a new list without " +
		"the continue parameter."
	inconsistentContinue string = "The provided continue parameter is too old " +
		"to display a consistent list result. You can start a new list without " +
		"the continue parameter, or use the continue token in this response to " +
		"retrieve the remainder of the results. Continuing with the provided " +
		"token results in an inconsistent list - objects that were created, " +
		"modified, or deleted between the time the first chunk was returned " +
		"and now may show up in the list."
)

func interpretListError(err error, paging bool, continueKey, keyPrefix string) error {
	switch {
	case err == etcdrpc.ErrCompacted:
		if paging {
			return handleCompactedErrorForPaging(continueKey, keyPrefix)
		}
		return errors.NewResourceExpired(expired)
	}
	return err
}

func handleCompactedErrorForPaging(continueKey, keyPrefix string) error {
	// continueToken.ResoureVersion=-1 means that the apiserver can
	// continue the list at the latest resource version. We don't use rv=0
	// for this purpose to distinguish from a bad token that has empty rv.
	newToken, err := storage.EncodeContinue(continueKey, keyPrefix, -1)
	if err != nil {
		utilruntime.HandleError(err)
		return errors.NewResourceExpired(continueExpired)
	}
	statusError := errors.NewResourceExpired(inconsistentContinue)
	statusError.ErrStatus.ListMeta.Continue = newToken
	return statusError
}

type StatusError struct {
	ErrStatus metav1.Status
}

// Error implements the Error interface.
func (e *StatusError) Error() string {
	return e.ErrStatus.Message
}

// Status allows access to e's status without having to know the detailed workings
// of StatusError.
func (e *StatusError) Status() metav1.Status {
	return e.ErrStatus
}

// DebugError reports extended info about the error to debug output.
func (e *StatusError) DebugError() (string, []interface{}) {
	if out, err := json.MarshalIndent(e.ErrStatus, "", "  "); err == nil {
		return "server response object: %s", []interface{}{string(out)}
	}
	return "server response object: %#v", []interface{}{e.ErrStatus}
}

var _ error = &StatusError{}

// NewBadRequest creates an error that indicates that the request is invalid and can not be processed.
func NewBadRequest(reason string) *StatusError {
	return &StatusError{metav1.Status{
		Status:  metav1.StatusFailure,
		Code:    http.StatusBadRequest,
		Reason:  metav1.StatusReasonBadRequest,
		Message: reason,
	}}
}
