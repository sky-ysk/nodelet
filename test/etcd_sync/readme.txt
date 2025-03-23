这里是一个用于跨域数据同步的小组件，需要使用infomer和apiserver所以放在了这里，使用的具体步骤如下：
1.按照apiserver部分文档的步骤启动apiserver
./_output/local/go/bin/apiserver
2.

http://broker.registry-svc.test.svc.clusterset.local:8081/forward

http://broker.registry-svc.test.svc.clusterset.local:3001/forward?target=<127.0.0.1:14399>

http://pve2.registry-svc.test.svc.clusterset.local:8081/forward


http://127.0.0.1:14399
http://pve2.registry-svc.test.svc.clusterset.local:3001/forward?target=127.0.0.1:14399


http://broker.registry-svc.test.svc.clusterset.local:3001/forward?target=<127.0.0.1:14399>

request.Header.Set("ClusterID", "broker")

request.Header.Set("ClusterID", "pve2")

172.100.0.109， 一个172.110.0.109
