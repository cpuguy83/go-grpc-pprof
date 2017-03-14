package server

import (
	"bytes"
	"os"
	"runtime"
	"runtime/pprof"
	"runtime/trace"
	"strings"

	gogotypes "github.com/gogo/protobuf/types"

	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"

	"github.com/cpuguy83/go-grpc-pprof/api"
)

// NewServer creates a new pprof server
func NewServer() api.PProfServiceServer {
	return server{}
}

type server struct{}

func (s server) Cmdline(ctx context.Context, req *api.CmdlineRequest) (*api.CmdlineResponse, error) {
	return &api.CmdlineResponse{
		Command: strings.Join(os.Args, "\x00"),
	}, nil
}

func (s server) CPUProfile(req *api.CPUProfileRequest, stream api.PProfService_CPUProfileServer) error {
	duration, err := gogotypes.DurationFromProto(req.Duration)
	if err != nil || duration == 0 {
		return grpc.Errorf(codes.InvalidArgument, "passed in duration is invalid: %v", req.Duration)
	}

	ctx, cancel := context.WithTimeout(stream.Context(), duration)
	defer cancel()

	if err := pprof.StartCPUProfile(api.NewChunkWriter(stream)); err != nil {
		return err
	}

	<-ctx.Done()
	pprof.StopCPUProfile()
	return nil
}

func (s server) Trace(req *api.TraceRequest, stream api.PProfService_TraceServer) error {
	duration, err := gogotypes.DurationFromProto(req.Duration)
	if err != nil || duration == 0 {
		return grpc.Errorf(codes.InvalidArgument, "passed in duration is invalid: %v", req.Duration)
	}

	ctx, cancel := context.WithTimeout(stream.Context(), duration)
	defer cancel()

	if err := trace.Start(api.NewChunkWriter(stream)); err != nil {
		return err
	}
	<-ctx.Done()
	trace.Stop()
	return nil
}

func (s server) Symbol(ctx context.Context, req *api.SymbolRequest) (*api.SymbolResponse, error) {
	f := runtime.FuncForPC(uintptr(req.Symbol))
	if f == nil {
		return nil, grpc.Errorf(codes.NotFound, "symbol %#x not found", req.Symbol)
	}

	return &api.SymbolResponse{
		Name:   f.Name(),
		Symbol: req.Symbol,
	}, nil
}

func (s server) Lookup(ctx context.Context, req *api.LookupRequest) (*api.LookupResponse, error) {
	p := pprof.Lookup(req.Name)
	if p == nil {
		return nil, grpc.Errorf(codes.NotFound, "could not find profile with name: %s", req.Name)
	}

	if req.Name == "heap" && req.GcBeforeHeap {
		runtime.GC()
	}

	var buf bytes.Buffer
	if err := p.WriteTo(&buf, int(req.Debug)); err != nil {
		return nil, err
	}

	return &api.LookupResponse{
		Data: buf.Bytes(),
	}, nil
}
