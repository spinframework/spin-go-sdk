module github.com/spinframework/spin-go-sdk/v3

go 1.25.5

require (
	github.com/bytecodealliance/wit-bindgen v0.0.0-00010101000000-000000000000
	github.com/julienschmidt/httprouter v1.3.0
)

replace github.com/bytecodealliance/wit-bindgen => github.com/bytecodealliance/wit-bindgen/crates/go/src/package v0.51.0
