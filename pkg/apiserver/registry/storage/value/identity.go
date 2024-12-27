package value

import (
	"bytes"
	"context"
	"fmt"
)

var (
	transformer      = identityTransformer{}
	encryptedPrefix  = []byte("k8s:enc:")
	errEncryptedData = fmt.Errorf("identity transformer tried to read encrypted data")
)

// identityTransformer performs no transformation on provided data, but validates
// that the data is not encrypted data during TransformFromStorage
type identityTransformer struct{}

// NewEncryptCheckTransformer returns an identityTransformer which returns an error
// on attempts to read encrypted data
func NewEncryptCheckTransformer() Transformer {
	return transformer
}

// TransformFromStorage returns the input bytes if the data is not encrypted
func (identityTransformer) TransformFromStorage(ctx context.Context, data []byte, dataCtx Context) ([]byte, bool, error) {
	// identityTransformer has to return an error if the data is encoded using another transformer.
	// JSON data starts with '{'. Protobuf data has a prefix 'k8s[\x00-\xFF]'.
	// Prefix 'k8s:enc:' is reserved for encrypted data on disk.
	if bytes.HasPrefix(data, encryptedPrefix) {
		return nil, false, errEncryptedData
	}
	return data, false, nil
}

// TransformToStorage implements the Transformer interface for identityTransformer
func (identityTransformer) TransformToStorage(ctx context.Context, data []byte, dataCtx Context) ([]byte, error) {
	return data, nil
}
