package http

import (
	"fmt"
	"io"

	"github.com/spinframework/spin-go-sdk/v3/internal/wasi/io/v0.2.0/streams"
)

type inputStreamReader struct {
	stream streams.InputStream
}

func (r inputStreamReader) Close() error {
	//noop
	return nil
}

func (r inputStreamReader) Read(p []byte) (n int, err error) {
	readResult := r.stream.Read(uint64(len(p)))
	if readResult.IsErr() {
		readErr := readResult.Err()
		if readErr.Closed() {
			return 0, io.EOF
		}
		return 0, fmt.Errorf("failed to read from InputStream %s", readErr.LastOperationFailed().ToDebugString())
	}

	readList := readResult.OK()
	copy(p, readList.Slice())
	return int(readList.Len()), nil
}

// create an io.Reader from the input stream
func NewReader(s streams.InputStream) io.Reader {
	return inputStreamReader{
		stream: s,
	}
}

func NewReadCloser(s streams.InputStream) io.ReadCloser {
	return inputStreamReader{
		stream: s,
	}
}
