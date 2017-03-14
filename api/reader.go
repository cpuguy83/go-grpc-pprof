package api

import (
	"bytes"
	"io"
)

// NewChunkReader creates an io.Reader from a ChunkReceiver.
// You can optionally pass in your own buffer to use or one will be created.
// The reader works to stitch back together Chunks (a protobuf encapsulated stream of bytes)
// as a single stream.
func NewChunkReader(recv ChunkReceiver, buf []byte) io.Reader {
	if buf != nil {
		buf = buf[0:]
	}
	return &chunkReader{r: recv, buf: bytes.NewBuffer(buf)}
}

// ChunkReceiver is when creating a chunk reader to be able to accept different
// streaming endpoints from the API server.
type ChunkReceiver interface {
	Recv() (*Chunk, error)
}

type chunkReader struct {
	buf *bytes.Buffer
	r   ChunkReceiver
}

func (r *chunkReader) Read(b []byte) (nr int, err error) {
	var chunk *Chunk
	for r.buf.Len() < len(b) {
		chunk, err = r.r.Recv()
		if chunk != nil {
			r.buf.Write(chunk.Chunk)
		}
		if err != nil {
			break
		}
	}
	return r.buf.Read(b)
}
