package preflight

import (
	"fmt"
	"math/rand"
	"net"
	"net/url"
	"time"
)

const connectionTimeout = 1 * time.Second

// etcd服务器列表
type EtcdConnection struct {
	ServerList []string
}

func (EtcdConnection) serverReachable(connURL *url.URL) bool {
	scheme := connURL.Scheme
	if scheme == "http" || scheme == "https" || scheme == "tcp" {
		scheme = "tcp"
	}
	if conn, err := net.DialTimeout(scheme, connURL.Host, connectionTimeout); err == nil {
		defer conn.Close()
		return true
	}
	return false
}

func parseServerURI(serverURI string) (*url.URL, error) {
	connURL, err := url.Parse(serverURI)
	if err != nil {
		return &url.URL{}, fmt.Errorf("unable to parse etcd url: %v", err)
	}
	return connURL, nil
}

// 检查是否服务器列表可达
func (con EtcdConnection) CheckEtcdServers() (done bool, err error) {
	// Attempt to reach every Etcd server randomly.
	serverNumber := len(con.ServerList)
	serverPerms := rand.Perm(serverNumber)
	for _, index := range serverPerms {
		host, err := parseServerURI(con.ServerList[index])
		if err != nil {
			return false, err
		}
		if con.serverReachable(host) {
			return true, nil
		}
	}
	return false, nil
}
