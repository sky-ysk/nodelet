package rmf

import (
	"log"
	"testing"
)

func TestRMF(t *testing.T) {
	done := make(chan struct{})

	go func() {
		// 构造一个测试用的request，供GetDeviceInfo使用
		request := Request{
			URL:    "http://192.168.1.225:8000",
			Fleets: "tinyRobot",
			Name:   "fixedArm",
			Params: make([]string, 0),
		}

		// 测试GetDeviceInfo
		t.Logf("----------rmf的测试结果----------\n")
		infos, err := GetDeviceInfo(request)
		if err != nil {
			log.Fatal(err)
		}

		// 打印每个属性
		for _, info := range infos.Infos {
			if len(info.SubProperty) > 0 {
				t.Logf("\tname:%s, type: %s, value: %s\n", info.Name, info.Type, info.Value)
				for _, subProperty := range info.SubProperty {
					t.Logf("\t\t|name: %s, type: %s, value: %s\n", subProperty.Name, subProperty.Type, subProperty.Value)
				}
			} else {
				t.Logf("\tname: %s, type: %s, value: %s\n", info.Name, info.Type, info.Value)
			}
		}

		done <- struct{}{}
	}()

	<-done
	close(done)
}
