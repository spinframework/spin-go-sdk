package pg

import (
	"database/sql/driver"
	"fmt"
	"time"

	pg "github.com/spinframework/spin-go-sdk/v3/imports/spin_postgres_4_2_0_postgres"
	wittypes "go.bytecodealliance.org/pkg/wit/types"
)

// | Go type                 | WIT (db-value)                                | Postgres type(s)             |
// |-------------------------|-----------------------------------------------|----------------------------- |
// | `bool`                  | boolean(bool)                                 | BOOL                         |
// | `int16`                 | int16(s16)                                    | SMALLINT, SMALLSERIAL, INT2  |
// | `int32`                 | int32(s32)                                    | INT, SERIAL, INT4            |
// | `int64`                 | int64(s64)                                    | BIGINT, BIGSERIAL, INT8      |
// | `float32`               | floating32(float32)                           | REAL, FLOAT4                 |
// | `float64`               | floating64(float64)                           | DOUBLE PRECISION, FLOAT8     |
// | `string`                | str(string)                                   | VARCHAR, CHAR(N), TEXT       |
// | `[]byte`                | binary(list\<u8\>)                            | BYTEA                        |
// | `Date`                  | date(tuple<s32, u8, u8>)                      | DATE                         |
// | `Time`                  | time(tuple<u8, u8, u8, u32>)                  | TIME                         |
// | `time.Time`             | datetime(tuple<s32, u8, u8, u8, u8, u8, u32>) | TIMESTAMP                    |
// | `time.Time`             | timestamp(s64)                                | BIGINT                       |
// | `UUID`                  | uuid(string)                                  | UUID                         |
// | `JSONB`                 | jsonb(list\<u8\>)                             | JSONB                        |
// | `Decimal`               | decimal(string)                               | NUMERIC                      |
// | `Int32Range`            | range-int32(...)                              | INT4RANGE                    |
// | `Int64Range`            | range-int64(...)                              | INT8RANGE                    |
// | `DecimalRange`          | range-decimal(...)                            | NUMERICRANGE                 |
// | `[]int32`               | array-int32(...)                              | INT4[]                       |
// | `[]int64`               | array-int64(...)                              | INT8[]                       |
// | `[]string`              | array-str(...)                                | TEXT[]                       |
// | `[]Decimal`             | array-decimal(...)                            | NUMERIC[]                    |
// | `Interval`              | interval(interval)                            | INTERVAL                     |

// Date represents a PostgreSQL date value.
type Date struct {
	Year  int
	Month time.Month
	Day   int
}

// Scan implements [sql.Scanner] so Date can be used as a scan destination.
func (d *Date) Scan(src any) error {
	switch src := src.(type) {
	case time.Time:
		d.Year = src.Year()
		d.Month = src.Month()
		d.Day = src.Day()
	case nil:
		*d = Date{}
	default:
		return fmt.Errorf("pg: cannot scan %T into *Date", src)
	}
	return nil
}

// Value implements [driver.Valuer] so Date can be used as a query parameter.
func (d Date) Value() (driver.Value, error) {
	return wittypes.Tuple3[int32, uint8, uint8]{
		F0: int32(d.Year),
		F1: uint8(d.Month),
		F2: uint8(d.Day),
	}, nil
}

// Time represents a PostgreSQL time value (time of day without date).
type Time struct {
	Hour       int
	Minute     int
	Second     int
	Nanosecond int
}

// Scan implements [sql.Scanner] so Time can be used as a scan destination.
func (t *Time) Scan(src any) error {
	switch src := src.(type) {
	case time.Time:
		t.Hour = src.Hour()
		t.Minute = src.Minute()
		t.Second = src.Second()
		t.Nanosecond = src.Nanosecond()
	case nil:
		*t = Time{}
	default:
		return fmt.Errorf("pg: cannot scan %T into *Time", src)
	}
	return nil
}

// Value implements [driver.Valuer] so Time can be used as a query parameter.
func (t Time) Value() (driver.Value, error) {
	return wittypes.Tuple4[uint8, uint8, uint8, uint32]{
		F0: uint8(t.Hour),
		F1: uint8(t.Minute),
		F2: uint8(t.Second),
		F3: uint32(t.Nanosecond),
	}, nil
}

// Interval represents a PostgreSQL interval value.
//
// PostgreSQL intervals have three components: months, days, and microseconds.
// Months and days are stored separately because the number of days in a month
// varies, and a day may have 23 or 25 hours due to daylight savings.
type Interval struct {
	Months int32
	Days   int32
	Micros int64
}

// Scan implements [sql.Scanner] so Interval can be used as a scan destination.
func (iv *Interval) Scan(src any) error {
	switch src := src.(type) {
	case Interval:
		*iv = src
	case nil:
		*iv = Interval{}
	default:
		return fmt.Errorf("pg: cannot scan %T into *Interval", src)
	}
	return nil
}

// Value implements [driver.Valuer] so Interval can be used as a query parameter.
func (iv Interval) Value() (driver.Value, error) {
	return pg.Interval{
		Micros: iv.Micros,
		Days:   iv.Days,
		Months: iv.Months,
	}, nil
}

// JSONB represents a PostgreSQL jsonb value.
type JSONB []byte

// Scan implements [sql.Scanner] so JSONB can be used as a scan destination.
func (j *JSONB) Scan(src any) error {
	switch src := src.(type) {
	case []byte:
		*j = make(JSONB, len(src))
		copy(*j, src)
	case pg.DbValue:
		if src.Tag() == pg.DbValueJsonb {
			data := src.Jsonb()
			*j = make(JSONB, len(data))
			copy(*j, data)
		}
	case nil:
		*j = nil
	default:
		return fmt.Errorf("pg: cannot scan %T into *JSONB", src)
	}
	return nil
}

// Value implements [driver.Valuer] so JSONB can be used as a query parameter.
func (j JSONB) Value() (driver.Value, error) {
	return pg.MakeDbValueJsonb([]byte(j)), nil
}

// UUID represents a PostgreSQL uuid value.
type UUID string

// Scan implements [sql.Scanner] so UUID can be used as a scan destination.
func (u *UUID) Scan(src any) error {
	switch src := src.(type) {
	case string:
		*u = UUID(src)
	case pg.DbValue:
		if src.Tag() == pg.DbValueUuid {
			uuid := src.Uuid()
			*u = UUID(uuid)
		}
	case nil:
		*u = ""
	default:
		return fmt.Errorf("pg: cannot scan %T into *UUID", src)
	}
	return nil
}

// Value implements [driver.Valuer] so UUID can be used as a query parameter.
func (u UUID) Value() (driver.Value, error) {
	return pg.MakeDbValueUuid(string(u)), nil
}

// Decimal represents a PostgreSQL numeric/decimal value.
//
// Values are stored as strings to preserve arbitrary precision. Use Decimal
// when exact numeric representation matters (e.g., monetary values).
type Decimal string

// Scan implements [sql.Scanner] so Decimal can be used as a scan destination.
func (d *Decimal) Scan(src any) error {
	switch src := src.(type) {
	case string:
		*d = Decimal(src)
	case nil:
		*d = ""
	default:
		return fmt.Errorf("pg: cannot scan %T into *Decimal", src)
	}
	return nil
}

// Value implements [driver.Valuer] so Decimal can be used as a query parameter.
func (d Decimal) Value() (driver.Value, error) {
	return pg.MakeDbValueDecimal(string(d)), nil
}

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

// Int64Range represents a PostgreSQL int8range value.
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

// DecimalRange represents a PostgreSQL numrange value.
//
// A nil Lower or Upper indicates an unbounded (infinite) bound.
// LowerInclusive and UpperInclusive specify whether each bound is included in the range.
type DecimalRange struct {
	Lower          *Decimal
	LowerInclusive bool
	Upper          *Decimal
	UpperInclusive bool
}

// Scan implements [sql.Scanner] so DecimalRange can be used as a scan destination.
func (r *DecimalRange) Scan(src any) error {
	if src == nil {
		*r = DecimalRange{}
		return nil
	}

	v, ok := src.(wittypes.Tuple2[
		wittypes.Option[wittypes.Tuple2[string, pg.RangeBoundKind]],
		wittypes.Option[wittypes.Tuple2[string, pg.RangeBoundKind]],
	])
	if !ok {
		return fmt.Errorf("pg: cannot scan %T into *DecimalRange", src)
	}

	*r = DecimalRange{}

	if v.F0.IsSome() {
		lb := v.F0.Some()
		val := Decimal(lb.F0)
		r.Lower = &val
		r.LowerInclusive = lb.F1 == pg.RangeBoundKindInclusive
	}

	if v.F1.IsSome() {
		ub := v.F1.Some()
		val := Decimal(ub.F0)
		r.Upper = &val
		r.UpperInclusive = ub.F1 == pg.RangeBoundKindInclusive
	}

	return nil
}

// Value implements [driver.Valuer] so DecimalRange can be used as a query parameter.
func (r DecimalRange) Value() (driver.Value, error) {
	var lower wittypes.Option[wittypes.Tuple2[string, pg.RangeBoundKind]]
	if r.Lower != nil {
		kind := pg.RangeBoundKindExclusive
		if r.LowerInclusive {
			kind = pg.RangeBoundKindInclusive
		}
		lower = wittypes.Some(wittypes.Tuple2[string, pg.RangeBoundKind]{F0: string(*r.Lower), F1: kind})
	} else {
		lower = wittypes.None[wittypes.Tuple2[string, pg.RangeBoundKind]]()
	}

	var upper wittypes.Option[wittypes.Tuple2[string, pg.RangeBoundKind]]
	if r.Upper != nil {
		kind := pg.RangeBoundKindExclusive
		if r.UpperInclusive {
			kind = pg.RangeBoundKindInclusive
		}
		upper = wittypes.Some(wittypes.Tuple2[string, pg.RangeBoundKind]{F0: string(*r.Upper), F1: kind})
	} else {
		upper = wittypes.None[wittypes.Tuple2[string, pg.RangeBoundKind]]()
	}

	return wittypes.Tuple2[
		wittypes.Option[wittypes.Tuple2[string, pg.RangeBoundKind]],
		wittypes.Option[wittypes.Tuple2[string, pg.RangeBoundKind]],
	]{F0: lower, F1: upper}, nil
}
