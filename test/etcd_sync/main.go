package main

import (
	"hit.edu/framework/test/etcd_sync/informer"
)

func main() {
	// 目前已经实现了域内的监视，只要把需要转发到的域的IP和端口写在这里，然后里面加上域和资源信息即可
	// TODO：把得到的消息转发出去
	informer.Synctest("http://127.0.0.1:14399")
}
