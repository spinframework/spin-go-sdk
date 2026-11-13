package export_wasi_http_0_3_2_handler

import (
	. "github.com/spinframework/spin-go-sdk/v4/imports/wasi_http_0_3_2_types"
	. "go.bytecodealliance.org/pkg/wit/types"
)

var Exports struct {
	Handle func(request *Request) Result[*Response, ErrorCode]
}

func Handle(request *Request) Result[*Response, ErrorCode] {
	return Exports.Handle(request)
}
