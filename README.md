## grpc/pprof for go

This implements a grpc service for requesting data from pprof.
You can think of this like `net/http/pprof` but for grpc.

Additionally includes an HTTP proxy for setting up an HTTP endpoint similar to
the ones provided by `net/http/pprof` but proxies to the GRPC service.
You can point `go tool pprof` at t this proxy and it should all just work.

See `cmd/` for examples on how to use these and to just generally test it.
