package http

import (
	"fmt"
	"io"

	. "github.com/spinframework/spin-go-sdk/v3/imports/wasi_http_0_3_0_rc_2026_03_15_types"
	. "go.bytecodealliance.org/pkg/wit/types"
)

// Assert `bodyReader` implements the required interfaces
var _ io.Reader = &bodyReader{}
var _ io.ReadCloser = &bodyReader{}

type bodyReader struct {
	stream   *StreamReader[uint8]
	trailers *FutureReader[Result[Option[*Fields], ErrorCode]]
}

func (self *bodyReader) Close() error {
	if self.stream != nil {
		self.stream.Drop()
	}
	if self.trailers != nil {
		self.trailers.Drop()
	}
	return nil
}

func (self *bodyReader) Read(p []byte) (n int, err error) {
	if self.stream.WriterDropped() {
		return 0, self.takeError()
	}

	count := self.stream.Read(p)
	if count == 0 && self.stream.WriterDropped() {
		return 0, self.takeError()
	}

	return int(count), nil
}

func (self *bodyReader) takeError() error {
	if self.trailers != nil {
		trailers := self.trailers.Read()
		self.trailers = nil
		if trailers.IsErr() {
			return fmt.Errorf("failed to read from HTTP body stream: %v", errorString(trailers.Err()))
		}
	}
	return io.EOF
}

// create an io.Reader from the input stream
func newReader(stream *StreamReader[uint8], trailers *FutureReader[Result[Option[*Fields], ErrorCode]]) *bodyReader {
	return &bodyReader{
		stream:   stream,
		trailers: trailers,
	}
}
