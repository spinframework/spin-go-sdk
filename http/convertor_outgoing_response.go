package http

import (
	"fmt"
	"net/http"

	. "github.com/spinframework/spin-go-sdk/v3/imports/wasi_http_0_3_0_rc_2026_03_15_types"
	. "go.bytecodealliance.org/pkg/wit/types"
)

// Assert that `responseWriter` implements the required interface
var _ http.ResponseWriter = &responseWriter{}

type responseWriter struct {
	// channel to which the response will be sent
	channel chan Result[*Response, ErrorCode]
	// stream to which the response body is being written
	stream *StreamWriter[uint8]
	// future which resolves to an error if there is a problem delivering the response body
	streamResult *FutureReader[Result[Unit, ErrorCode]]
	// headers to send
	headers http.Header
	// status code to send
	statusCode int
}

func (self *responseWriter) Header() http.Header {
	return self.headers
}

func (self *responseWriter) Write(buf []byte) (int, error) {
	err := self.send()
	if err != nil {
		return 0, err
	}

	count := self.stream.Write(buf)

	if count == 0 && self.stream.ReaderDropped() {
		return 0, self.takeError()
	}

	return int(count), nil
}

func (self *responseWriter) WriteHeader(statusCode int) {
	self.statusCode = statusCode
}

func (self *responseWriter) close() {
	if self.stream != nil {
		self.stream.Drop()
	}
	if self.streamResult != nil {
		self.streamResult.Drop()
	}
}

func (self *responseWriter) send() error {
	channel := self.channel

	if channel == nil {
		return nil
	} else {
		self.channel = nil
	}

	fields, err := toWasiHeaders(self.headers)
	if err != nil {
		return err
	}

	tx, rx := MakeStreamU8()
	self.stream = tx

	response, send := ResponseNew(
		fields,
		Some(rx),
		trailersFuture(), // TODO: support trailers
	)
	self.streamResult = send

	response.SetStatusCode(uint16(self.statusCode))

	channel <- Ok[*Response, ErrorCode](response)

	return nil
}

func (r *responseWriter) takeError() error {
	if r.streamResult != nil {
		streamResult := r.streamResult.Read()
		r.streamResult = nil
		if streamResult.IsErr() {
			return fmt.Errorf(
				"failed to read from HTTP body stream: %v",
				errorString(streamResult.Err()),
			)
		}
	}
	return nil
}

func newHttpResponseWriter() *responseWriter {
	return &responseWriter{
		channel:    make(chan Result[*Response, ErrorCode]),
		headers:    http.Header{},
		statusCode: 200,
	}
}
