package main

import (
	"log"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/cpuguy83/go-grpc-pprof/api"
	httpproxy "github.com/cpuguy83/go-grpc-pprof/http"
	"google.golang.org/grpc"
)

func main() {
	sock := os.Args[1]
	conn, err := grpc.Dial(sock, grpc.WithInsecure(), grpc.WithBlock(), grpc.WithDialer(
		func(addr string, timeout time.Duration) (net.Conn, error) {
			return net.DialTimeout("unix", addr, timeout)
		}))
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	client := api.NewPProfServiceClient(conn)
	http.Handle("/debug/pprof/", httpproxy.NewProxy(client))
	http.ListenAndServe("127.0.0.1:8080", http.DefaultServeMux)
}
