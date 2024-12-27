// 这个文件的作用是支持分页查询，处理continue token
// 目前基本上是从apiserver中照搬过来的，可能需要进行部分修改,比如所有的apiVersion都是"meta.k8s.io/v1"
package storage

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"path"
	"strings"
)

var (
	ErrInvalidStartRV             = errors.New("continue key is not valid: incorrect encoded start resourceVersion (version meta.k8s.io/v1)")
	ErrEmptyStartKey              = errors.New("continue key is not valid: encoded start key empty (version meta.k8s.io/v1)")
	ErrGenericInvalidKey          = errors.New("continue key is not valid")
	ErrUnrecognizedEncodedVersion = errors.New("continue key is not valid: server does not recognize this encoded version")
)

// continueToken 指示continue token的状态
type continueToken struct {
	APIVersion      string `json:"v"`
	ResourceVersion int64  `json:"rv"`
	StartKey        string `json:"start"` //继续查询的起始键
}

// DecodeContinue 解析continueValue提取继续查询的起始键、资源版本号等信息
func DecodeContinue(continueValue, keyPrefix string) (fromKey string, rv int64, err error) {
	data, err := base64.RawURLEncoding.DecodeString(continueValue)
	if err != nil {
		return "", 0, fmt.Errorf("%w: %v", ErrGenericInvalidKey, err)
	}
	var c continueToken
	if err := json.Unmarshal(data, &c); err != nil {
		return "", 0, fmt.Errorf("%w: %v", ErrGenericInvalidKey, err)
	}
	switch c.APIVersion {
	case "meta.k8s.io/v1":
		if c.ResourceVersion == 0 {
			return "", 0, ErrInvalidStartRV
		}
		if len(c.StartKey) == 0 {
			return "", 0, ErrEmptyStartKey
		}
		key := c.StartKey
		if !strings.HasPrefix(key, "/") {
			key = "/" + key
		}
		cleaned := path.Clean(key)
		if cleaned != key {
			return "", 0, fmt.Errorf("%w: %v", ErrGenericInvalidKey, c.StartKey)
		}
		return keyPrefix + cleaned[1:], c.ResourceVersion, nil
	default:
		return "", 0, fmt.Errorf("%w %v", ErrUnrecognizedEncodedVersion, c.APIVersion)
	}
}

// EncodeContinue 将继续查询的状态编码成continue token
func EncodeContinue(key, keyPrefix string, resourceVersion int64) (string, error) {
	nextKey := strings.TrimPrefix(key, keyPrefix)
	if nextKey == key {
		return "", fmt.Errorf("unable to encode next field: the key and key prefix do not match")
	}
	out, err := json.Marshal(&continueToken{APIVersion: "meta.k8s.io/v1", ResourceVersion: resourceVersion, StartKey: nextKey})
	if err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(out), nil
}

// PrepareContinueToken 在有更多要返回数据的情况下，生成continue token
func PrepareContinueToken(keyLastItem, keyPrefix string, resourceVersion int64, itemsCount int64, hasMoreItems bool, opts ListOptions) (string, *int64, error) {
	var remainingItemCount *int64
	var continueValue string
	var err error

	if hasMoreItems {
		// 分页查询从当前的最后一个键的下一个开始
		continueValue, err = EncodeContinue(keyLastItem+"\x00", keyPrefix, resourceVersion)
		if err != nil {
			return "", remainingItemCount, err
		}
		// 只有在查询条件为空时返回剩余数据项的数量
		if opts.Predicate.Empty() {
			remainingItems := itemsCount - opts.Predicate.Limit
			remainingItemCount = &remainingItems
		}
	}
	return continueValue, remainingItemCount, err
}
