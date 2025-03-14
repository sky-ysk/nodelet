package workflow

// TODO：该部分后面可能放到全局调度器中
// TODO: 检查新创建的Workflow，为Workflow创建Task,Group和Action子资源，并且为其分配对应的ID
// TODO: 由全局调度器，执行任务的分组算法和跨域调度算法，决定将任务部署在本地还是远端
// TODO: 对于本地任务，本地调度器会开始调度任务
// TODO：对于全局任务，由全局调度器为任务打上标签，并且由跨域通信的组件将相关数据同步到其他域

// TODO: 可以绕过Workflow,直接创建Task
