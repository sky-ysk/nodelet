package storage

import (
	"hit.edu/framework/pkg/apimachinery/errors"
	"hit.edu/framework/pkg/apimachinery/runtime/schema"
	"hit.edu/framework/pkg/apiserver/registry/storage"
)

// InterpretListError converts a generic error on a retrieval
// operation into the appropriate API error.
func InterpretListError(err error, qualifiedResource schema.GroupResource) error {
	switch {
	case storage.IsNotFound(err):
		return errors.NewNotFound(qualifiedResource, "")
	case storage.IsUnreachable(err), storage.IsRequestTimeout(err):
		return errors.NewServerTimeout(qualifiedResource, "list", 2) // TODO: make configurable or handled at a higher level
	case storage.IsInternalError(err):
		return errors.NewInternalError(err)
	default:
		return err
	}
}

// InterpretGetError converts a generic error on a retrieval
// operation into the appropriate API error.
func InterpretGetError(err error, qualifiedResource schema.GroupResource, name string) error {
	switch {
	case storage.IsNotFound(err):
		return errors.NewNotFound(qualifiedResource, name)
	case storage.IsUnreachable(err):
		return errors.NewServerTimeout(qualifiedResource, "get", 2) // TODO: make configurable or handled at a higher level
	case storage.IsInternalError(err):
		return errors.NewInternalError(err)
	default:
		return err
	}
}

// InterpretCreateError converts a generic error on a create
// operation into the appropriate API error.
func InterpretCreateError(err error, qualifiedResource schema.GroupResource, name string) error {
	switch {
	case storage.IsExist(err):
		return errors.NewAlreadyExists(qualifiedResource, name)
	case storage.IsUnreachable(err):
		return errors.NewServerTimeout(qualifiedResource, "create", 2) // TODO: make configurable or handled at a higher level
	case storage.IsInternalError(err):
		return errors.NewInternalError(err)
	default:
		return err
	}
}

// InterpretUpdateError converts a generic error on an update
// operation into the appropriate API error.
func InterpretUpdateError(err error, qualifiedResource schema.GroupResource, name string) error {
	switch {
	case storage.IsConflict(err), storage.IsExist(err), storage.IsInvalidObj(err):
		return errors.NewConflict(qualifiedResource, name, err)
	case storage.IsUnreachable(err):
		return errors.NewServerTimeout(qualifiedResource, "update", 2) // TODO: make configurable or handled at a higher level
	case storage.IsNotFound(err):
		return errors.NewNotFound(qualifiedResource, name)
	case storage.IsInternalError(err):
		return errors.NewInternalError(err)
	default:
		return err
	}
}

// InterpretDeleteError converts a generic error on a delete
// operation into the appropriate API error.
func InterpretDeleteError(err error, qualifiedResource schema.GroupResource, name string) error {
	switch {
	case storage.IsNotFound(err):
		return errors.NewNotFound(qualifiedResource, name)
	case storage.IsUnreachable(err):
		return errors.NewServerTimeout(qualifiedResource, "delete", 2) // TODO: make configurable or handled at a higher level
	case storage.IsConflict(err), storage.IsExist(err), storage.IsInvalidObj(err):
		return errors.NewConflict(qualifiedResource, name, err)
	case storage.IsInternalError(err):
		return errors.NewInternalError(err)
	default:
		return err
	}
}

// InterpretWatchError converts a generic error on a watch
// operation into the appropriate API error.
func InterpretWatchError(err error, resource schema.GroupResource, name string) error {
	switch {
	case storage.IsInvalidError(err):
		invalidError, _ := err.(storage.InvalidError)
		return errors.NewInvalid(schema.GroupKind{Group: resource.Group, Kind: resource.Resource}, name, invalidError.Errs)
	case storage.IsInternalError(err):
		return errors.NewInternalError(err)
	default:
		return err
	}
}
