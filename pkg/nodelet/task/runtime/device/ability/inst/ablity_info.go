package inst

import (
	"encoding/json"
	"errors"
	"fmt"
	"hit.edu/framework/pkg/component-base/logs"
	"io"
	"net/http"
	"time"
)

type AbilityInstance struct {
	Id           string       `json:"id"`
	Kind         string       `json:"kind"`
	MetaData     MetaData     `json:"metadata"`
	Owner        Owner        `json:"owner"`
	RunInfo      RunInfo      `json:"runInfo"`
	Shares       interface{}  `json:"sharers"`
	Spec         Spec         `json:"spec"`
	Status       interface{}  `json:"status"`
	Tag          Tag          `json:"tag"`
	Subabilities []SubAbility `json:"subabilities"`
}

type MetaData struct {
	Labels interface{} `json:"labels"`
	Name   string      `json:"name"`
}

type Owner struct {
	AbilityInstanceId string `json:"abilityInstanceId"`
	Position          string `json:"position"`
}

type RunInfo struct {
	LastConnect    int    `json:"lastConnect"`
	LastUpdate     int    `json:"lastUpdate"`
	LifeCycleState string `json:"lifecycleState"`
}

type Spec struct {
	AbilityName       string            `json:"abilityName"`
	ActivityCondition ActivityCondition `json:"activityCondition"`
	Config            interface{}       `json:"config"`
	Package           string            `json:"package"`
	SubAbilities      []Spec            `json:"subabilities"`
	Position          string            `json:"position"`
	Priority          int               `json:"priority"`
	Version           string            `json:"version"`
}

type ActivityCondition struct {
	JQ string `json:"jq"`
}
type Tag struct {
	Parent string `json:"parent"`
	Source string `json:"source"`
}

type Request struct {
	Url     string
	Payload string
}

type SubAbility struct {
	Id       string `json:"id"`
	Position string `json:"position"`
}

func NewPostRequest(url string, payload string) *Request {
	return &Request{Url: url, Payload: payload}
}
func NewGetRequest(url string) *Request {
	return &Request{Url: url}
}

func GetAbilityInstances(url string) ([]AbilityInstance, error) {
	requestForGet := NewGetRequest(fmt.Sprintf("%s/api/cr", url))
	// 创建HTTP client
	client := &http.Client{
		Timeout: 5 * time.Second,
	}
	logs.Info("publish get request\n")
	response, err := client.Get(requestForGet.Url)
	if err != nil {
		logs.Error("get response error: %v\n", err)
		return []AbilityInstance{}, err
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		logs.Error("response code is %v\n", response.StatusCode)
		return []AbilityInstance{}, errors.New(response.Status)
	}
	responseBody, err := io.ReadAll(response.Body)
	if err != nil {
		return []AbilityInstance{}, err
	}
	var abilityInstances []AbilityInstance
	err = json.Unmarshal(responseBody, &abilityInstances)
	if err != nil {
		logs.Error("unmarshal response error: %v\n", err)
		return []AbilityInstance{}, err
	}
	logs.Info("unmarshal response successfully\n")
	return abilityInstances, nil
}

// FindIdByAbilityName 根据能力名字寻找uuid
func FindIdByAbilityName(abilityName string, abilityInstances []AbilityInstance) (string, error) {
	for _, abilityInstance := range abilityInstances {
		if abilityInstance.Spec.AbilityName == abilityName {
			return abilityInstance.Id, nil
		}
		if abilityInstance.Subabilities != nil {
			for index, subability := range abilityInstance.Spec.SubAbilities {
				if subability.AbilityName == abilityName {
					return abilityInstance.Subabilities[index].Id, nil
				}
			}
		}
	}
	return "", errors.New(abilityName + " is not found")
}
