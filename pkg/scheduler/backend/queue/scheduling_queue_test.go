package queue

import (
	"context"
	apis "hit.edu/framework/pkg/apis/cores"
	"testing"
	"time"
)

// TODO 部署，数据库-- resource let
//TODO bind client - go

// 测试成功，调度队列的基础部分可以正常运行：往Pending里面加任务，然后不断从Pending里面把可以执行的任务放到Active
func TestPriorityQueue(t *testing.T) {
	NewPriorityQueue()
	q := NewPriorityQueue()
	ctx := context.Background()
	group1 := &apis.Group{
		Spec: apis.GroupSpec{
			Name: "g1",
		},
	}
	group2 := &apis.Group{
		Spec: apis.GroupSpec{
			Name: "g2",
		},
	}
	q.Run(ctx)
	q.AddToPending(ctx, group1)
	q.AddToPending(ctx, group2)
	time.Sleep(time.Duration(10) * time.Second)
	t.Log("success run TestPriorityQueue_Init test")
}
