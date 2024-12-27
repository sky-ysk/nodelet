package storagebackend

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	TestServer "hit.edu/framework/pkg/apiserver/registry/storagebackend/server"
)

// 测试错误信息记录
func Test_atomicLastError(t *testing.T) {
	aError := &atomicLastError{err: fmt.Errorf("initial error")}

	aError.Store(errors.New("updated error"), time.Time{})
	err := aError.Load()
	if err.Error() != "updated error" {
		t.Fatalf("Expected: \"updated error\" got: %s", err.Error())
	}

	now := time.Now()
	aError.Store(errors.New("now error"), now)
	err = aError.Load()
	if err.Error() != "now error" {
		t.Fatalf("Expected: \"now error\" got: %s", err.Error())
	}

	past := now.Add(-5 * time.Second)
	aError.Store(errors.New("past error"), past)
	err = aError.Load()
	if err.Error() != "now error" {
		t.Fatalf("Expected: \"now error\" got: %s", err.Error())
	}
}

// 测试客户端的创建
func Test_newETCD3Client(t *testing.T) {
	client, ListenClientUrls := TestServer.RunEtcd(t, nil)
	var clientUrl []string
	for _, u := range ListenClientUrls {
		clientUrl = append(clientUrl, u.String())
	}
	TransConfig := TransportConfig{
		ServerList:    clientUrl,
		KeyFile:       "",
		CertFile:      "",
		TrustedCAFile: "",
	}
	testclient, _ := NewETCD3Client(TransConfig)
	client.Put(context.TODO(), "name", "test")
	resp, _ := client.Get(context.TODO(), "name")
	var value1, value2 string
	for _, ev := range resp.Kvs {
		value1 = string(ev.Value)
	}
	resp1, _ := testclient.Get(context.TODO(), "name")
	for _, ev := range resp1.Kvs {
		value2 = string(ev.Value)
	}
	if value1 != value2 {
		t.Fatalf("value of name got different")
	}

}
