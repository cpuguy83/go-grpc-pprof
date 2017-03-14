package http

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"net/http"
	"path"
	"strconv"
	"strings"
	"time"

	gogotypes "github.com/gogo/protobuf/types"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"

	"github.com/cpuguy83/go-grpc-pprof/api"
)

var respCodes = map[codes.Code]int{
	codes.Unknown:         http.StatusInternalServerError,
	codes.InvalidArgument: http.StatusBadRequest,
	codes.NotFound:        http.StatusNotFound,
	codes.AlreadyExists:   http.StatusConflict,
}

func mkError(w http.ResponseWriter, err error) {
	if code, exists := respCodes[grpc.Code(err)]; exists {
		http.Error(w, err.Error(), code)
		return
	}
	http.Error(w, err.Error(), http.StatusInternalServerError)
}

// NewProxy returns a pprof http proxy which proxies requests to a grpc
// service.
// It returns an http.Handler that you can use to route requests to.
// For instance `http.Handle("/debug/pprof/", NewHTTPProxy(client)`
func NewProxy(c api.PProfServiceClient) http.Handler {
	return &httpProxy{c}
}

type httpProxy struct {
	c api.PProfServiceClient
}

func (p *httpProxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch path.Base(r.URL.Path) {
	case "profile":
		p.CPUProfile(w, r)
	case "cmdline":
		p.Cmdline(w, r)
	case "symbol":
		p.Symbol(w, r)
	case "trace":
		p.Trace(w, r)
	default:
		p.Lookup(w, r)
	}
}

func (p *httpProxy) Cmdline(w http.ResponseWriter, req *http.Request) {
	res, err := p.c.Cmdline(req.Context(), nil)
	if err != nil {
		mkError(w, err)
		return
	}
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	fmt.Fprintf(w, res.Command)
}

func (p *httpProxy) CPUProfile(w http.ResponseWriter, req *http.Request) {
	sec, _ := strconv.ParseInt(req.FormValue("seconds"), 10, 64)
	if sec == 0 {
		sec = 30
	}

	stream, err := p.c.CPUProfile(req.Context(), &api.CPUProfileRequest{
		Duration: gogotypes.DurationProto(time.Duration(sec) * time.Second),
	})
	if err != nil {
		mkError(w, err)
		return
	}

	r := api.NewChunkReader(stream)
	w.Header().Set("Content-Type", "application/octet-stream")
	io.Copy(makeFlusher(w), r)
}

func (p *httpProxy) Trace(w http.ResponseWriter, req *http.Request) {
	sec, _ := strconv.ParseInt(req.FormValue("seconds"), 10, 64)
	if sec == 0 {
		sec = 30
	}

	stream, err := p.c.Trace(req.Context(), &api.TraceRequest{
		Duration: gogotypes.DurationProto(time.Duration(sec) * time.Second),
	})
	if err != nil {
		mkError(w, err)
		return
	}

	r := api.NewChunkReader(stream)
	w.Header().Set("Content-Type", "application/octet-stream")
	io.Copy(makeFlusher(w), r)
}

// much of this is taken from net/http/pprof with some the actual runtime call
// changed to call out to the grpc service.
func (p *httpProxy) Symbol(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")

	// We have to read the whole POST body before
	// writing any output. Buffer the output here.
	var buf bytes.Buffer

	// We don't know how many symbols we have, but we
	// do have symbol information. Pprof only cares whether
	// this number is 0 (no symbols available) or > 0.
	fmt.Fprintf(&buf, "num_symbols: 1\n")

	var b *bufio.Reader
	if req.Method == "POST" {
		b = bufio.NewReader(req.Body)
	} else {
		b = bufio.NewReader(strings.NewReader(req.URL.RawQuery))
	}

	var symbols []uint64

	for {
		word, err := b.ReadSlice('+')
		if err == nil {
			word = word[0 : len(word)-1] // trim +
		}
		pc, _ := strconv.ParseUint(string(word), 0, 64)
		if pc != 0 {
			symbols = append(symbols, pc)

			// Call GRPC instead of runtime
			// This is where this code differs from net/http/pprof
			res, err := p.c.Symbol(req.Context(), &api.SymbolRequest{
				Symbol: pc,
			})
			if err == nil {
				fmt.Fprintf(&buf, "%#x %s\n", res.Symbol, res.Name)
			}

		}
		if err != nil {
			if err != io.EOF {
				fmt.Fprintf(&buf, "reading request: %v\n", err)
			}
			break
		}
	}

	w.Write(buf.Bytes())
}

func (p *httpProxy) Lookup(w http.ResponseWriter, req *http.Request) {
	debug, _ := strconv.Atoi(req.FormValue("debug"))
	gc, _ := strconv.Atoi(req.FormValue("gc"))
	res, err := p.c.Lookup(req.Context(), &api.LookupRequest{
		Name:         strings.TrimPrefix(req.URL.Path, "/debug/pprof/"),
		Debug:        int32(debug),
		GcBeforeHeap: gc > 0,
	})

	if err != nil {
		mkError(w, err)
		return
	}

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Write(res.Data)
}
