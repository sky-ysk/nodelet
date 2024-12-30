package queue

import (
	"context"
	"fmt"
	apis "hit.edu/framework/pkg/apis/cores"
	"testing"
)

// TODO 部署，数据库-- resource let
//TODO bind client - go

// 测试成功，调度队列的基础部分可以正常运行：往Pending里面加任务，然后不断从Pending里面把可以执行的任务放到Active
func TestPriorityQueue(t *testing.T) {
	q := NewPriorityQueue()
	ctx := context.Background()
	t.Log("initial priority queue")

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
	info, _ := q.Pop(ctx)
	fmt.Println(info.Group.Spec.Name)
	//time.Sleep(time.Duration(3) * time.Second)
	t.Log("success run TestPriorityQueue_Init test")
}

func TestReadyQueue(t *testing.T) {
	q := NewPriorityQueue()
	q.readyQ.lock.Lock()
	q.readyQ.cond.Wait()
	//q := newReadyQueue()
	//q.lock.Lock()
	//q.cond.Wait()
}
