package grpc_server

//import (
//	"context"
//	"google.golang.org/grpc"
//	"google.golang.org/grpc/codes"
//	"google.golang.org/grpc/status"
//	"hit.edu/resourcelet/pkg/component-base/logs"
//	pb "hit.edu/resourcelet/pkg/nodelet/common/taskInteractionProto"
//	"hit.edu/resourcelet/pkg/nodelet/exporter/taskexporter/entity"
//	EM "hit.edu/resourcelet/pkg/nodelet/exporter/taskexporter/entity/enitymanager"
//	"net"
//)
//
//// 这里相当于是部署器和切换器的核心代码区域 负责接收Controller发来的指令
//type Server struct {
//	Port string
//	pb.UnimplementedTaskNoticeServer
//}
//
//func NewServer(port string) *Server {
//	return &Server{Port: port}
//}
//
//func (s *Server) StartServer() bool {
//	listen, err := net.Listen("tcp", ":"+s.Port)
//	if err != nil {
//		logs.Error("listen grpc-server failed:%v", err)
//		return false
//	}
//	grpcServer := grpc.NewServer()             //调用的是grpc官方的方法
//	pb.RegisterTaskNoticeServer(grpcServer, s) //必须通过引用注册
//
//	//启动服务
//	err = grpcServer.Serve(listen)
//	if err != nil {
//		logs.Error("fail to start grpc-server:%v", err)
//		return false
//	}
//	logs.Info("start grpc-server success")
//	return true
//}
//
//func (s *Server) StartTask(ctx context.Context, req *pb.StartTaskRequest) (*pb.TaskResponse, error) {
//	task := convertProtoToGoTask(req.GetTask())
//	logs.Info("Taskexporter received TaskInfo, put into ready queue")
//	EM.GetInstance().AddToPending(task.GetId(), task)
//	return &pb.TaskResponse{IsSuccess: true}, nil
//}
//func (s *Server) GetTaskStatus(context.Context, *pb.TaskRequest) (*pb.TaskStatusResponse, error) {
//	logs.Info("Checking status for task")
//	//TODO 获取任务状态 直接调用与任务建立的rpc方法即可
//	return nil, status.Errorf(codes.Unimplemented, "method GetTaskStatus not implemented")
//}
//func (s *Server) StopTask(context.Context, *pb.TaskRequest) (*pb.TaskResponse, error) {
//	//TODO 暂停任务
//
//	return nil, status.Errorf(codes.Unimplemented, "method StopTask not implemented")
//}
//func (s *Server) RestartTask(context.Context, *pb.TaskRequest) (*pb.TaskResponse, error) {
//	//TODO 重启任务
//	return nil, status.Errorf(codes.Unimplemented, "method RestartTask not implemented")
//}
//func (s *Server) StopExporter(context.Context, *pb.TaskEmpty) (*pb.TaskResponse, error) {
//	//TODO 关闭Taskexporter进程
//	return nil, status.Errorf(codes.Unimplemented, "method StopExporter not implemented")
//}
//func convertProtoToGoTask(t *pb.Task) *entity.Task {
//	return entity.NewInputTask(t.TaskId, t.SourceIp, t.TaskType, t.DeviceType, t.Port)
//}
