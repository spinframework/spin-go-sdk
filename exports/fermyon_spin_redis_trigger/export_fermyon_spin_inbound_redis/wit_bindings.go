package export_fermyon_spin_inbound_redis

import (
	redis_types "github.com/spinframework/spin-go-sdk/v3/imports/fermyon_spin_redis_types"
	wit "go.bytecodealliance.org/pkg/wit/types"
)

var Exports struct {
	Handle func(message []byte) wit.Result[wit.Unit, redis_types.Error]
}

func HandleMessage(message []byte) wit.Result[wit.Unit, redis_types.Error] {
	return Exports.Handle(message)
}
