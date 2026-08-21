package main

import (
	"agentops/agentservice/handlers"
	"agentops/agentservice/services"
	"log"
	"net"
	"google.golang.org/grpc"
)
type GRPCServer struct{
	addr string
}

func NewGRPCServer(addr string)(*GRPCServer){
	return &GRPCServer{addr:addr}
}

func (g *GRPCServer) Run() error{
	lis,err := net.Listen("tcp",g.addr)

	if err!=nil{
		log.Println("Failed to listen on", g.addr, ":", err)
		return err
	}
	server := grpc.NewServer()
	agent_service:=services.NewAgentService()
	handlers.NewAgentService(server,agent_service)
	log.Println("Starting server on", g.addr)
	return server.Serve(lis)
}
