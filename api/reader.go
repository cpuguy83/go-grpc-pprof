package api

import (
	"bytes"
	"io"
)

// NewChunkReader creates an io.Reader from a ChunkReceiver.
// The reader works to stitch back together Chunks (a protobuf encapsulated stream of bytes)
// as a single stream.
func NewChunkReader(recv ChunkReceiver) io.Reader {
	return &chunkReader{r: recv}
}

// ChunkReceiver is when creating a chunk reader to be able to accept different
// streaming endpoints from the API server.
type ChunkReceiver interface {
	Recv() (*Chunk, error)
}

type chunkReader struct {
	buf io.Reader
	r   ChunkReceiver
}

func (r *chunkReader) Read(b []byte) (nr int, err error) {
	max := len(b)
	var read int
	var endStream bool

	for nr < max {
		if r.buf != nil {
			read, err = r.buf.Read(b[nr:max])
			nr += read
			if nr == max {
				return
			}
			if endStream && err != nil {
				return
			}
		}

		if r.buf == nil || (err != nil && !endStream) {
			chunk, rerr := r.r.Recv()
			if chunk != nil {
				if r.buf != nil {
					r.buf = io.MultiReader(r.buf, bytes.NewReader(chunk.Chunk))
				} else {
					r.buf = bytes.NewReader(chunk.Chunk)
				}
			}
			if rerr != nil {
				endStream = true
			}
		}
	}

	return
}
