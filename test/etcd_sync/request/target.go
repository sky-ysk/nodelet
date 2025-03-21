package request

import (
	"fmt"

	"hit.edu/framework/pkg/client-go/rest"
)

type Config struct {
	Targets []Target
	Client  *rest.Config
}

type Target struct {
	ClusterID   string //域名
	ServiceIP   string //跨域服务的URL
	ServicePort int    //跨域服务的端口
	ServerIP    string //跨域同步的URL
	ServerPort  int    //跨域同步的端口
	//EtcdIP      string //域内etcdIP
	//EtcdPort    int    //域内etcd端口
}

func PrintTargets(config Config) {
	for _, target := range config.Targets {
		fmt.Println(CreateURL(target))
	}

}
func AddToTargets(config *Config, clusterID, serviceIP string, servicePort int, serverIP string, serverPort int) error {
	for i, target := range config.Targets {
		//已有集群，进行信息更新
		if target.ClusterID == clusterID {
			config.Targets[i] = Target{
				ClusterID:   clusterID,
				ServiceIP:   serviceIP,
				ServicePort: servicePort,
				ServerIP:    serverIP,
				ServerPort:  serverPort,
			}
			fmt.Println("Updated existing target:", config.Targets[i])
			return nil
		}
	}

	// 新增条目
	newTarget := Target{
		ClusterID:   clusterID,
		ServiceIP:   serviceIP,
		ServicePort: servicePort,
		ServerIP:    serverIP,
		ServerPort:  serverPort,
	}
	config.Targets = append(config.Targets, newTarget)
	return nil
}

func UpdateTarget(config *Config, clusterID, serviceIP string, servicePort int, serverIP string, serverPort int) error {
	for i, target := range config.Targets {
		if target.ClusterID == clusterID {
			config.Targets[i] = Target{
				ClusterID:   clusterID,
				ServiceIP:   serviceIP,
				ServicePort: servicePort,
				ServerIP:    serverIP,
				ServerPort:  serverPort,
			}
			return nil
		}
	}
	return fmt.Errorf("clusterID %s not found", clusterID)
}

func DeleteTarget(config *Config, clusterID string) error {
	for i, target := range config.Targets {
		if target.ClusterID == clusterID {
			config.Targets = append(config.Targets[:i], config.Targets[i+1:]...)
			return nil
		}
	}
	return fmt.Errorf("clusterID %s not found", clusterID)
}

func CreateURL(target Target) string {
	url := fmt.Sprintf("http://%s.%s:%d/forward?target=%s:%d", target.ClusterID, target.ServiceIP, target.ServicePort, target.ServerIP, target.ServerPort)
	return url
}
