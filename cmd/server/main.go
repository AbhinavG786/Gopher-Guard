package main

import (
	"log/slog"
	"net"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"
	"net/http"
	"github.com/AbhinavG786/Gopher-Guard/internal/engine"
	mygrpc "github.com/AbhinavG786/Gopher-Guard/internal/grpc"
	"github.com/AbhinavG786/Gopher-Guard/internal/grpc/pb"
	"github.com/joho/godotenv"
	"google.golang.org/grpc"
	myraft "github.com/AbhinavG786/Gopher-Guard/internal/raft" 
	"github.com/hashicorp/raft"
)

func main(){
	_=godotenv.Load()
	port:=getEnv("PORT","50051")
	adminPort:=getEnv("ADMIN_PORT","8080")
	janitorInterval:=getEnvAsInt("JANITOR_INTERVAL_SEC",60)
	maxWindow:=getEnvAsInt("MAX_WINDOW_SEC",3600)
	nodeID:=getEnv("NODE_ID","node-1")
	isBootstrap:=getEnv("BOOTSTRAP","false")=="true"
	raftBindAddr := getEnv("RAFT_BIND", "127.0.0.1:7000") 
	raftDir := getEnv("RAFT_DIR", "./raft-data")
	
	logger:=slog.New(slog.NewJSONHandler(os.Stdout,nil))
	slog.SetDefault(logger)

	slog.Info("Starting Gopher-Guard",slog.String("port",port))

	rateLimiterEngine:=engine.NewSlidingWindow()
	rateLimiterEngine.StartJanitor(time.Duration(janitorInterval)*time.Second,time.Duration(maxWindow)*time.Second)
	slog.Info("Janitor Started",slog.Int("Janitor_interval_sec",janitorInterval))

	lis,err:= net.Listen("tcp","127.0.0.1:"+port)
	if err!=nil{
		slog.Error("Failed to listen on port", slog.String("error", err.Error()))
		os.Exit(1)
	}
	slog.Info("Server is listening on", slog.String("address", lis.Addr().String()))

	limiterFSM:=&engine.LimiterFSM{
		Engine: rateLimiterEngine,
	}

	raftNode,err:=myraft.SetupRaft(nodeID,raftBindAddr,raftDir,limiterFSM,isBootstrap)
	if err!=nil{
		slog.Error("Failed to set up Raft node", slog.String("error", err.Error()))
		os.Exit(1)
	}

	rateLimiterEngine.RaftNode=raftNode

	grpcServer:=grpc.NewServer()

	limiterServer:=&mygrpc.Server{
		Limiter: rateLimiterEngine,
	}
	
	pb.RegisterRateLimiterServer(grpcServer,limiterServer)

	http.HandleFunc("/join",func (w http.ResponseWriter,r *http.Request)  {
		if raftNode.State()!= raft.Leader {
			http.Error(w,"Not the leader",http.StatusBadGateway)
			return
		}

		joinNodeID:=r.URL.Query().Get("id")
		joinNodeAddr:=r.URL.Query().Get("addr")

		if joinNodeID=="" || joinNodeAddr==""{
			http.Error(w,"Missing id or addr parameter",http.StatusBadRequest)
			return
		}

		future:=raftNode.AddVoter(raft.ServerID(joinNodeID),raft.ServerAddress(joinNodeAddr),0,0)
		if err:=future.Error(); err!=nil{
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		slog.Info("Node joined cluster", slog.String("node_id", joinNodeID), slog.String("node_addr", joinNodeAddr))
		w.WriteHeader(http.StatusOK)		
	})

	go func(){
		slog.Info("Admin API listening", slog.String("port", adminPort))
		if err:=http.ListenAndServe(":"+adminPort,nil);err!=nil{
			slog.Error("Admin API crashed", slog.String("error", err.Error()))
		}
	}()

	stopChan:=make(chan os.Signal,1)
	signal.Notify(stopChan,os.Interrupt,syscall.SIGTERM)
	go func(){
	if err:=grpcServer.Serve(lis); err!=nil{
		slog.Error("gRPC server crashed", slog.String("error", err.Error()))
	}
}()
	sig:=<-stopChan
	slog.Info("Shutting down server", slog.String("signal", sig.String()))
	grpcServer.GracefulStop()
	slog.Info("Server stopped gracefully")
}

func getEnv(key string, fallback string) string{
	if value,exists:=os.LookupEnv(key); exists{
		return value
	}
	return fallback
}

func getEnvAsInt(key string,fallback int) int{
	strValue:=getEnv(key,"")
	if value,err:=strconv.Atoi(strValue); err==nil{
		return value
	}
	return fallback
}