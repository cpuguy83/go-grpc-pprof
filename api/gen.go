//go:generate protoc -I ../vendor:../vendor/github.com/gogo/protobuf/protobuf:. --gofast_out=plugins=grpc,import_path=github.com/cpuguy83/go-grpc-pprof/api,Mgoogle/protobuf/duration.proto=github.com/gogo/protobuf/types:. pprof.proto

package api
