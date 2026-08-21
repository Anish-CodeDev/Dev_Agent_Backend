package handlers

import (
	"context"
	"google.golang.org/grpc"
	types "agentops/agentservice/types"
	agent "agentops/common"
)

type AgentHandler struct{
	service types.AgentService
	agent.UnimplementedManageAgentOpsServer
}

func (h *AgentHandler) ExecuteCommands(ctx context.Context, in *agent.ExecuteCommandsRequest)(*agent.ExecuteCommandsResponse,error){
	err := h.service.ExecuteCommands(ctx,in)
	if err !=nil{
		return nil,err
	}
	return &agent.ExecuteCommandsResponse{Status: "Successful"},nil
}

func (h *AgentHandler) CreateFiles(ctx context.Context,in *agent.CreateFileRequest)(*agent.CreateFileResponse,error){
	err:=h.service.CreateFiles(ctx,in)
	if err!=nil{
		return nil,err
	}
	return &agent.CreateFileResponse{
		Status: "Successful",
	},nil
}

func (h *AgentHandler) ViewFile(ctx context.Context,in *agent.ViewFileRequest)(*agent.ViewFileResponse,error){
	err,code:= h.service.ViewFiles(ctx,&agent.ViewFileRequest{
		Path: in.Path,
	})
	if err!=nil{
		return nil,err
	}
	return &agent.ViewFileResponse{
		Code: code,
	},nil
}
func NewAgentService(grpc *grpc.Server, services types.AgentService){
	grpcHandler := &AgentHandler{
		service: services,
	}
	agent.RegisterManageAgentOpsServer(grpc,grpcHandler)
}