package pg

import (
	"database/sql/driver"
	"fmt"

	pg "github.com/spinframework/spin-go-sdk/v3/imports/spin_postgres_4_2_0_postgres"
	wittypes "go.bytecodealliance.org/pkg/wit/types"
)

// Int32Range represents a PostgreSQL int4range value.
//
// A nil Lower or Upper indicates an unbounded (infinite) bound.
// LowerInclusive and UpperInclusive specify whether each bound is included in the range.
type Int32Range struct {
	Lower          *int32
	LowerInclusive bool
	Upper          *int32
	UpperInclusive bool
}

// Scan implements [sql.Scanner] so Int32Range can be used as a scan destination.
func (r *Int32Range) Scan(src any) error {
	if src == nil {
		*r = Int32Range{}
		return nil
	}

	v, ok := src.(wittypes.Tuple2[
		wittypes.Option[wittypes.Tuple2[int32, pg.RangeBoundKind]],
		wittypes.Option[wittypes.Tuple2[int32, pg.RangeBoundKind]],
	])
	if !ok {
		return fmt.Errorf("pg: cannot scan %T into *Int32Range", src)
	}

	*r = Int32Range{}

	if v.F0.IsSome() {
		lb := v.F0.Some()
		val := lb.F0
		r.Lower = &val
		r.LowerInclusive = lb.F1 == pg.RangeBoundKindInclusive
	}

	if v.F1.IsSome() {
		ub := v.F1.Some()
		val := ub.F0
		r.Upper = &val
		r.UpperInclusive = ub.F1 == pg.RangeBoundKindInclusive
	}

	return nil
}

// Value implements [driver.Valuer] so Int32Range can be used as a query parameter.
func (r Int32Range) Value() (driver.Value, error) {
	var lower wittypes.Option[wittypes.Tuple2[int32, pg.RangeBoundKind]]
	if r.Lower != nil {
		kind := pg.RangeBoundKindExclusive
		if r.LowerInclusive {
			kind = pg.RangeBoundKindInclusive
		}
		lower = wittypes.Some(wittypes.Tuple2[int32, pg.RangeBoundKind]{F0: *r.Lower, F1: kind})
	} else {
		lower = wittypes.None[wittypes.Tuple2[int32, pg.RangeBoundKind]]()
	}

	var upper wittypes.Option[wittypes.Tuple2[int32, pg.RangeBoundKind]]
	if r.Upper != nil {
		kind := pg.RangeBoundKindExclusive
		if r.UpperInclusive {
			kind = pg.RangeBoundKindInclusive
		}
		upper = wittypes.Some(wittypes.Tuple2[int32, pg.RangeBoundKind]{F0: *r.Upper, F1: kind})
	} else {
		upper = wittypes.None[wittypes.Tuple2[int32, pg.RangeBoundKind]]()
	}

	return wittypes.Tuple2[
		wittypes.Option[wittypes.Tuple2[int32, pg.RangeBoundKind]],
		wittypes.Option[wittypes.Tuple2[int32, pg.RangeBoundKind]],
	]{F0: lower, F1: upper}, nil
}

// Int64Range represents a PostgreSQL int4range value.
//
// A nil Lower or Upper indicates an unbounded (infinite) bound.
// LowerInclusive and UpperInclusive specify whether each bound is included in the range.
type Int64Range struct {
	Lower          *int64
	LowerInclusive bool
	Upper          *int64
	UpperInclusive bool
}

// Scan implements [sql.Scanner] so Int64Range can be used as a scan destination.
func (r *Int64Range) Scan(src any) error {
	if src == nil {
		*r = Int64Range{}
		return nil
	}

	v, ok := src.(wittypes.Tuple2[
		wittypes.Option[wittypes.Tuple2[int64, pg.RangeBoundKind]],
		wittypes.Option[wittypes.Tuple2[int64, pg.RangeBoundKind]],
	])
	if !ok {
		return fmt.Errorf("pg: cannot scan %T into *Int64Range", src)
	}

	*r = Int64Range{}

	if v.F0.IsSome() {
		lb := v.F0.Some()
		val := lb.F0
		r.Lower = &val
		r.LowerInclusive = lb.F1 == pg.RangeBoundKindInclusive
	}

	if v.F1.IsSome() {
		ub := v.F1.Some()
		val := ub.F0
		r.Upper = &val
		r.UpperInclusive = ub.F1 == pg.RangeBoundKindInclusive
	}

	return nil
}

// Value implements [driver.Valuer] so Int64Range can be used as a query parameter.
func (r Int64Range) Value() (driver.Value, error) {
	var lower wittypes.Option[wittypes.Tuple2[int64, pg.RangeBoundKind]]
	if r.Lower != nil {
		kind := pg.RangeBoundKindExclusive
		if r.LowerInclusive {
			kind = pg.RangeBoundKindInclusive
		}
		lower = wittypes.Some(wittypes.Tuple2[int64, pg.RangeBoundKind]{F0: *r.Lower, F1: kind})
	} else {
		lower = wittypes.None[wittypes.Tuple2[int64, pg.RangeBoundKind]]()
	}

	var upper wittypes.Option[wittypes.Tuple2[int64, pg.RangeBoundKind]]
	if r.Upper != nil {
		kind := pg.RangeBoundKindExclusive
		if r.UpperInclusive {
			kind = pg.RangeBoundKindInclusive
		}
		upper = wittypes.Some(wittypes.Tuple2[int64, pg.RangeBoundKind]{F0: *r.Upper, F1: kind})
	} else {
		upper = wittypes.None[wittypes.Tuple2[int64, pg.RangeBoundKind]]()
	}

	return wittypes.Tuple2[
		wittypes.Option[wittypes.Tuple2[int64, pg.RangeBoundKind]],
		wittypes.Option[wittypes.Tuple2[int64, pg.RangeBoundKind]],
	]{F0: lower, F1: upper}, nil
}
