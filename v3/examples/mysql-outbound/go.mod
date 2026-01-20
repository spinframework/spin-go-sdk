module github.com/spinframework/spin-go-sdk/v3/examples/mysql-outbound

go 1.25.5

require github.com/spinframework/spin-go-sdk/v3 v3.0.0

require (
	github.com/bytecodealliance/wit-bindgen v0.0.0-00010101000000-000000000000 // indirect
	github.com/julienschmidt/httprouter v1.3.0 // indirect
)

replace github.com/spinframework/spin-go-sdk/v3 => ../../

replace github.com/bytecodealliance/wit-bindgen => github.com/bytecodealliance/wit-bindgen/crates/go/src/package v0.51.0
