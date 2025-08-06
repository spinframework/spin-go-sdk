module github.com/http_go

go 1.24

require github.com/spinframework/spin-go-sdk/v3 v3.0.0

require (
	github.com/julienschmidt/httprouter v1.3.0 // indirect
	go.bytecodealliance.org/cm v0.2.2 // indirect
)

replace github.com/spinframework/spin-go-sdk/v3 => ../../
