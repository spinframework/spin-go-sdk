module github.com/spinframework/spin-go-sdk/v3/examples/grpc

go 1.25.5

require github.com/spinframework/spin-go-sdk/v3 v3.0.0

require google.golang.org/grpc v1.83.2

require google.golang.org/protobuf v1.36.11

require (
	github.com/apparentlymart/go-userdirs v0.0.0-20200915174352-b0c018a67c13 // indirect
	github.com/bytecodealliance/componentize-go v0.4.0 // indirect
	github.com/gofrs/flock v0.13.0 // indirect
	go.bytecodealliance.org/pkg v0.2.3 // indirect
	golang.org/x/net v0.58.0 // indirect
	golang.org/x/sys v0.47.0 // indirect
	golang.org/x/text v0.41.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20260526163538-3dc84a4a5aaa // indirect
)

replace github.com/spinframework/spin-go-sdk/v3 => ../../

tool github.com/bytecodealliance/componentize-go
