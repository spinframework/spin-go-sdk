package http

import (
	"fmt"
	"io"

	streams "github.com/spinframework/spin-go-sdk/v3/imports/wasi_io_0_2_0_streams"
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
		if readErr.Tag() == streams.StreamErrorClosed {
			return 0, io.EOF
		}
		return 0, fmt.Errorf("failed to read from InputStream %s", readErr.LastOperationFailed().ToDebugString())
	}

	readList := readResult.Ok()
	copy(p, readList)
	return len(readList), nil
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
