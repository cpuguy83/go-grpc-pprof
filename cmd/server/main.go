package main

import (
	"net"
	"os"

	"github.com/Sirupsen/logrus"
	"github.com/cpuguy83/go-grpc-pprof/api"
	"github.com/cpuguy83/go-grpc-pprof/server"
	"golang.org/x/net/context"

	"google.golang.org/grpc"
)

func main() {
	sock := "pprof.sock"
	if len(os.Args) > 1 {
		sock = os.Args[1]
	}
	os.Remove(sock)

	l, err := net.Listen("unix", sock)
	if err != nil {
		logrus.Fatal(err)
	}

	srv := grpc.NewServer(grpc.UnaryInterceptor(withLogs), grpc.StreamInterceptor(withStreamLogs))
	api.RegisterPProfServiceServer(srv, server.NewServer())
	logrus.SetLevel(logrus.DebugLevel)

	srv.Serve(l)
}

func withLogs(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
	logrus.WithField("req", req).WithField("url", info.FullMethod).Debug("Start")
	resp, err = handler(ctx, req)
	logrus.WithField("req", req).WithField("url", info.FullMethod).Debug("Start")
	return
}

func withStreamLogs(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
	logrus.WithField("req", srv).WithField("url", info.FullMethod).Debug("Start")
	err := handler(srv, ss)
	logrus.WithField("req", srv).WithField("url", info.FullMethod).Debug("Start")
	return err
}
