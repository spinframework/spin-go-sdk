package http

import (
	"fmt"
	"net/http"
	"slices"

	wasi "github.com/spinframework/spin-go-sdk/v3/imports/wasi_http_0_3_0_rc_2026_03_15_types"
	wit "go.bytecodealliance.org/pkg/wit/types"
)

// Assert that `responseWriter` implements the required interface
var _ http.ResponseWriter = &responseWriter{}

type responseWriter struct {
	// channel to which the response will be sent
	channel chan wit.Result[*wasi.Response, wasi.ErrorCode]
	// stream to which the response body is being written
	stream *wit.StreamWriter[uint8]
	// future which resolves to an error if there is a problem delivering the response body
	streamResult *wit.FutureReader[wit.Result[wit.Unit, wasi.ErrorCode]]
	// headers to send
	headers http.Header
	// status code to send
	statusCode       int
	trailersTx       *wit.FutureWriter[wit.Result[wit.Option[*wasi.Fields], wasi.ErrorCode]]
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

func (self *responseWriter) Flush() {
}

func (self *responseWriter) writeTrailers() error {
	if self.trailersTx == nil {
		return nil
	}

	declared := self.headers.Values("Trailer")
	collected := make(http.Header)
	for headerName, headerVals := range self.headers {
		if slices.Contains(declared, headerName) {
			collected[headerName] = headerVals
		}
	}

	if len(collected) > 0 {
		wasiTrailers, err := toWasiHeaders(collected)
		if err != nil {
			return err
		}
		self.trailersTx.Write(wit.Ok[wit.Option[*wasi.Fields], wasi.ErrorCode](wit.Some(wasiTrailers)))
	} else {
		self.trailersTx.Write(wit.Ok[wit.Option[*wasi.Fields], wasi.ErrorCode](wit.None[*wasi.Fields]()))
	}

	self.trailersTx = nil
	return nil
}

func (self *responseWriter) close() error {
	err := self.writeTrailers()
	if err != nil {
		return err
	}

	if self.stream != nil {
		self.stream.Drop()
	}
	if self.streamResult != nil {
		self.streamResult.Drop()
	}
	if self.trailersTx != nil {
		self.trailersTx.Drop()
	}

	return nil
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

	tx, rx := wasi.MakeStreamU8()
	self.stream = tx

	trailersTx, trailersRx := wasi.MakeFutureResultOptionFieldsErrorCode()
	self.trailersTx = trailersTx

	response, send := wasi.ResponseNew(
		fields,
		wit.Some(rx),
		trailersRx,
	)
	self.streamResult = send

	response.SetStatusCode(uint16(self.statusCode))

	channel <- wit.Ok[*wasi.Response, wasi.ErrorCode](response)

	return nil
}

func (r *responseWriter) takeError() error {
	if r.streamResult != nil {
		streamResult := r.streamResult.Read()
		r.streamResult = nil
		if streamResult.IsErr() {
			return fmt.Errorf(
				"failed to read from HTTP body stream: %s",
				errorString(streamResult.Err()),
			)
		}
	}
	return nil
}

func newHttpResponseWriter() *responseWriter {
	return &responseWriter{
		channel:    make(chan wit.Result[*wasi.Response, wasi.ErrorCode]),
		headers:    http.Header{},
		statusCode: 200,
	}
}
