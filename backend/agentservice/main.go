package main

func main(){
server:=NewGRPCServer(":9000")
_ = server.Run()	
}