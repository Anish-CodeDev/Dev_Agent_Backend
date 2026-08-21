package main
import (
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"log"
)

func NewGRPCClient(addr string)*grpc.ClientConn{
	conn,err := grpc.NewClient(addr,grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err!=nil{
		log.Fatalf("failed to connect: %v", err)
	}
	return conn
}

func main(){
	server:=NewHttpServer(":5000")
	server.Run()
}
