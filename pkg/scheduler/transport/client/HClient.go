package client

import (
	"bytes"
	"encoding/json"
	"golang.org/x/net/http2"
	"hit.edu/framework/pkg/component-base/logs"
	"hit.edu/framework/pkg/scheduler/transport"
	"io"
	"net/http"
	"net/url"
	"sync"
	"time"
)

// 调度器的Http Client
type HClient struct {
	client *http.Client
}

// 定义一个私有变量来保存单例实例
var hClientInstance *HClient

// 定义互斥锁，用于保证并发安全
var hClientMutex sync.Mutex

// 双锁检测的单例 拿到Client实例
func GetHClient() *HClient {
	if hClientInstance == nil {
		hClientMutex.Lock()
		defer hClientMutex.Unlock()
		if hClientInstance == nil {
			hClientInstance = &HClient{
				client: &http.Client{
					Timeout: time.Second * 60,
					Transport: &http2.Transport{
						AllowHTTP: true, // 允许非加密的HTTP/2连接（测试环境可用，生产环境建议使用TLS加密）
					},
				},
			}
		}
	}
	return hClientInstance
}

func (hc *HClient) SendScoreRequest(request transport.ScoreRequest) transport.ScoreResponse {
	jsonData, err := json.Marshal(request)
	if err != nil {
		logs.Fatal(err)
		return transport.ScoreResponse{
			BaseResp: transport.BaseResponse{
				Code:   "400",
				Reason: err.Error(),
			},
		}
	}
	httpReq := http.Request{
		Method: "POST",
		URL: &url.URL{
			Host: request.Host,
			//TODO path定一下
			Path: "/score",
		},
	}
	httpReq.Body = io.NopCloser(bytes.NewBuffer(jsonData))
	httpRes, err := hc.client.Do(&httpReq)
	if err != nil {
		logs.Fatal(err)
		return transport.ScoreResponse{}
	}
	defer httpRes.Body.Close()
	body, err := io.ReadAll(httpRes.Body)
	if err != nil {
		logs.Fatal(err)
		return transport.ScoreResponse{}
	}
	var data transport.ScoreRespData
	err = json.Unmarshal(body, &data)
	if err != nil {
		logs.Fatal(err)
		return transport.ScoreResponse{}
	}
	return transport.ScoreResponse{
		BaseResp: transport.BaseResponse{},
		Data:     data,
	}
}
