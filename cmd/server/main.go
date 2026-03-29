package main;

import (
	"log"
	"net"
	"google.golang.org/grpc"
	"github.com/AbhinavG786/Gopher-Guard/internal/grpc/pb"
	mygrpc "github.com/AbhinavG786/Gopher-Guard/internal/grpc"
)

func main(){
	lis,err:= net.Listen("tcp",":50051")
	if err!=nil{
		log.Fatalf("Failed to listen: %v",err)
	}
	log.Printf("Server is listening on %v",lis.Addr())

	grpcServer:=grpc.NewServer()
	limiterServer:=&mygrpc.Server{}
	pb.RegisterRateLimiterServer(grpcServer,limiterServer)
	if err:=grpcServer.Serve(lis); err!=nil{
		log.Fatalf("Failed to serve: %v",err)
	}
}