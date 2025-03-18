package pool

import (
	"fmt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"sync"
	"time"
)

type ConnectionPool struct {
	pool   sync.Map // key=address, value=*grpc.ClientConn
	locker sync.Mutex
}

func NewConnectionPool() *ConnectionPool {
	return &ConnectionPool{
		pool:   sync.Map{},
		locker: sync.Mutex{},
	}
}

func (cp *ConnectionPool) GetConn(address string) (*grpc.ClientConn, error) {
	// 双重检查锁定模式
	if conn, ok := cp.pool.Load(address); ok {
		return conn.(*grpc.ClientConn), nil
	}

	cp.locker.Lock()
	defer cp.locker.Unlock()

	if conn, ok := cp.pool.Load(address); ok {
		return conn.(*grpc.ClientConn), nil
	}

	conn, err := grpc.Dial(address,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithTimeout(5*time.Second))
	if err != nil {
		return nil, err
	}

	cp.pool.Store(address, conn)
	return conn, nil
}

// connection_pool.go 增加带重试的GetConnWithRetry方法
func (cp *ConnectionPool) GetConnWithRetry(address string, maxRetries int, backoff time.Duration) (*grpc.ClientConn, error) {
	var conn *grpc.ClientConn
	var err error

	for i := 0; i < maxRetries; i++ {
		conn, err = cp.GetConn(address)
		if err == nil {
			return conn, nil
		}

		if i < maxRetries-1 {
			time.Sleep(backoff)
			backoff = time.Duration(float64(backoff) * 1.5) // 指数退避
		}
	}
	return nil, fmt.Errorf("after %d retries: %v", maxRetries, err)
}

// 修改 GetConn 实现健康检查
//func (cp *ConnectionPool) GetConn(address string) (*grpc.ClientConn, error) {
//	if conn, ok := cp.pool.Load(address); ok {
//		if conn.(*grpc.ClientConn).GetState() == connectivity.Ready {
//			return conn.(*grpc.ClientConn), nil
//		}
//	}
//
//	cp.mu.Lock()
//	defer cp.mu.Unlock()
//
//	// 双重检查
//	if conn, ok := cp.pool.Load(address); ok {
//		return conn.(*grpc.ClientConn), nil
//	}
//
//	conn, err := grpc.Dial(address,
//		grpc.WithTransportCredentials(insecure.NewCredentials()),
//		grpc.WithKeepaliveParams(keepalive.ClientParameters{
//			Time:    30 * time.Second,
//			Timeout: 10 * time.Second,
//		}))
//	if err != nil {
//		return nil, err
//	}
//
//	cp.pool.Store(address, conn)
//	return conn, nil
//}
