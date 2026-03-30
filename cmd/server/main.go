package main

import (
	"log/slog"
	"net"
	"time"
	"os"
	"os/signal"
	"syscall"
	"strconv"
	"github.com/joho/godotenv"
	"github.com/AbhinavG786/Gopher-Guard/internal/engine"
	mygrpc "github.com/AbhinavG786/Gopher-Guard/internal/grpc"
	"github.com/AbhinavG786/Gopher-Guard/internal/grpc/pb"
	"google.golang.org/grpc"
)

func main(){
	_=godotenv.Load()
	port:=getEnv("PORT","50051")
	janitorInterval:=getEnvAsInt("JANITOR_INTERVAL_SEC",60)
	maxWindow:=getEnvAsInt("MAX_WINDOW_SEC",3600)
	
	logger:=slog.New(slog.NewJSONHandler(os.Stdout,nil))
	slog.SetDefault(logger)

	slog.Info("Starting Gopher-Guard",slog.String("port",port))

	rateLimiterEngine:=engine.NewSlidingWindow()
	rateLimiterEngine.StartJanitor(time.Duration(janitorInterval)*time.Second,time.Duration(maxWindow)*time.Second)
	slog.Info("Janitor Started",slog.Int("Janitor_interval_sec",janitorInterval))

	lis,err:= net.Listen("tcp",":"+port)
	if err!=nil{
		slog.Error("Failed to listen on port", slog.String("error", err.Error()))
		os.Exit(1)
	}
	slog.Info("Server is listening on", slog.String("address", lis.Addr().String()))

	grpcServer:=grpc.NewServer()

	limiterServer:=&mygrpc.Server{
		Limiter: rateLimiterEngine,
	}
	
	pb.RegisterRateLimiterServer(grpcServer,limiterServer)

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