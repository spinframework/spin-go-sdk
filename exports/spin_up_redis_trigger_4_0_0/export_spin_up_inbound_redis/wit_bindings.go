package export_fermyon_spin_inbound_redis

import (
	redis "github.com/spinframework/spin-go-sdk/v3/imports/spin_redis_3_0_0_redis"
	wit "go.bytecodealliance.org/pkg/wit/types"
)

var Exports struct {
	Handle func(message []byte) wit.Result[wit.Unit, redis.Error]
}

func HandleMessage(message []byte) wit.Result[wit.Unit, redis.Error] {
	return Exports.Handle(message)
}
