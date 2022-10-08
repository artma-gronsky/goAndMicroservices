package main

import (
	"context"
	"fmt"
	"google.golang.org/grpc"
	"log"
	"logger-service/data"
	"logger-service/logs"
	"net"
)

type LogServer struct {
	logs.UnimplementedLogServiceServer
	Models data.Models
}

func (l *LogServer) WriteLog(_ context.Context, req *logs.LogRequest) (*logs.LogResponse, error) {
	input := req.GetLogEntry()

	// write the log
	logEntry := data.LogEntry{
		Data: input.Data,
	}

	err := l.Models.LogEntry.Insert(logEntry)

	if err != nil {
		return &logs.LogResponse{Result: "Failed"}, err
	}

	return &logs.LogResponse{Result: "Logged!"}, nil
}

func (app *Config) gRPCListen() {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", gRpcPort))

	if err != nil {
		log.Fatalf("Failed for listen grpc: %v", err)
		panic(err)
	}

	s := grpc.NewServer()

	logs.RegisterLogServiceServer(s, &LogServer{Models: app.Models})

	log.Printf("gRPC server started on port: %d", gRpcPort)

	if err = s.Serve(lis); err != nil {
		log.Fatalf("Failed for listen grpc: %v", err)
		panic(err)
	}
}
