// Package inbound_redis provides a handler for inbound Redis messages

package inbound_redis

import (
	incominghandler "github.com/spinframework/spin-go-sdk/v3/redis_internal/export_fermyon_spin_inbound_redis"
	redis_types "github.com/spinframework/spin-go-sdk/v3/redis_internal/fermyon_spin_redis_types"
	_ "github.com/spinframework/spin-go-sdk/v3/redis_internal/wit_exports"
	wit_dir "github.com/spinframework/spin-go-sdk/v3/wit"
	wit "go.bytecodealliance.org/pkg/wit/types"
)

// force wit files to be shipped with sdk dependency
var _ = wit_dir.Wit

func Handle(handle func(message []byte) error) {
	incominghandler.Exports.Handle = func(message []byte) wit.Result[wit.Unit, redis_types.Error] {
		err := handle(message)
		if err == nil {
			return wit.Err[wit.Unit, redis_types.Error](redis_types.ErrorError)
		} else {
			return wit.Ok[wit.Unit, redis_types.Error](wit.Unit{})
		}
	}
}
