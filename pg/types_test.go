package pg

import (
	"database/sql/driver"
	"testing"
	"time"

	pg "github.com/spinframework/spin-go-sdk/v3/imports/spin_postgres_4_2_0_postgres"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	wittypes "go.bytecodealliance.org/pkg/wit/types"
)

func ptr[T any](v T) *T { return &v }

func TestJSONB_Scan(t *testing.T) {
	tests := []struct {
		name    string
		src     any
		want    JSONB
		wantErr bool
	}{{
		name: "from []byte",
		src:  []byte(`{"key":"value"}`),
		want: JSONB(`{"key":"value"}`),
	}, {
		name: "from DbValue",
		src:  pg.MakeDbValueJsonb([]byte(`[1,2,3]`)),
		want: JSONB(`[1,2,3]`),
	}, {
		name: "nil src",
		src:  nil,
		want: nil,
	}, {
		name:    "invalid src type",
		src:     42,
		wantErr: true,
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var j JSONB
			err := j.Scan(tt.src)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, j)
		})
	}
}

func TestJSONB_Value(t *testing.T) {
	j := JSONB(`{"key":"value"}`)
	got, _ := j.Value()
	assert.Equal(t, driver.Value(pg.MakeDbValueJsonb([]byte(`{"key":"value"}`))), got)
}

func TestJSONB_RoundTrip(t *testing.T) {
	original := JSONB(`{"nested":{"a":1}}`)
	val, _ := original.Value()

	var recovered JSONB
	recovered.Scan(val)
	assert.Equal(t, original, recovered)
}

func TestDate_Scan(t *testing.T) {
	tests := []struct {
		name    string
		src     any
		want    Date
		wantErr bool
	}{{
		name: "from time.Time",
		src:  time.Date(2024, time.March, 15, 0, 0, 0, 0, time.UTC),
		want: Date{Year: 2024, Month: time.March, Day: 15},
	}, {
		name: "nil src",
		src:  nil,
		want: Date{},
	}, {
		name:    "invalid src type",
		src:     "2024-03-15",
		wantErr: true,
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var d Date
			err := d.Scan(tt.src)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, d)
		})
	}
}

func TestDate_Value(t *testing.T) {
	d := Date{Year: 2024, Month: time.March, Day: 15}
	got, _ := d.Value()
	assert.Equal(t, driver.Value(wittypes.Tuple3[int32, uint8, uint8]{
		F0: 2024, F1: 3, F2: 15,
	}), got)
}

func TestDate_RoundTrip(t *testing.T) {
	original := Date{Year: 2024, Month: time.December, Day: 25}
	witVal, _ := original.Value()

	tuple := witVal.(wittypes.Tuple3[int32, uint8, uint8])
	asTime := time.Date(int(tuple.F0), time.Month(tuple.F1), int(tuple.F2), 0, 0, 0, 0, time.UTC)

	var recovered Date
	recovered.Scan(asTime)
	assert.Equal(t, original, recovered)
}

func TestTime_Scan(t *testing.T) {
	tests := []struct {
		name    string
		src     any
		want    Time
		wantErr bool
	}{{
		name: "from time.Time",
		src:  time.Date(0, 1, 1, 14, 30, 45, 123456789, time.UTC),
		want: Time{Hour: 14, Minute: 30, Second: 45, Nanosecond: 123456789},
	}, {
		name: "nil src",
		src:  nil,
		want: Time{},
	}, {
		name:    "invalid src type",
		src:     "14:30:45",
		wantErr: true,
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var tm Time
			err := tm.Scan(tt.src)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, tm)
		})
	}
}

func TestTime_Value(t *testing.T) {
	original := Time{Hour: 14, Minute: 30, Second: 45, Nanosecond: 123456789}
	witVal, _ := original.Value()
	assert.Equal(t, driver.Value(wittypes.Tuple4[uint8, uint8, uint8, uint32]{
		F0: 14, F1: 30, F2: 45, F3: 123456789,
	}), witVal)
}

func TestTime_RoundTrip(t *testing.T) {
	original := Time{Hour: 23, Minute: 59, Second: 59, Nanosecond: 999000000}
	witVal, _ := original.Value()

	// Simulate what toRow does: convert the WIT tuple to time.Time
	tuple := witVal.(wittypes.Tuple4[uint8, uint8, uint8, uint32])
	asTime := time.Date(0, 1, 1, int(tuple.F0), int(tuple.F1), int(tuple.F2), int(tuple.F3), time.UTC)

	var recovered Time
	require.NoError(t, recovered.Scan(asTime))
	assert.Equal(t, original, recovered)
}

func TestInterval_Scan(t *testing.T) {
	tests := []struct {
		name    string
		src     any
		want    Interval
		wantErr bool
	}{{
		name: "full interval",
		src:  Interval{Months: 14, Days: 3, Micros: 7200000000},
		want: Interval{Months: 14, Days: 3, Micros: 7200000000},
	}, {
		name: "micros only",
		src:  Interval{Micros: 3600000000},
		want: Interval{Micros: 3600000000},
	}, {
		name: "nil src",
		src:  nil,
		want: Interval{},
	}, {
		name:    "invalid src type",
		src:     "1 year 2 months",
		wantErr: true,
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var iv Interval
			err := iv.Scan(tt.src)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, iv)
		})
	}
}

func TestInterval_Value(t *testing.T) {
	iv := Interval{Months: 14, Days: 3, Micros: 7200000000}
	got, _ := iv.Value()
	assert.Equal(t, driver.Value(pg.Interval{
		Micros: 7200000000, Days: 3, Months: 14,
	}), got)
}

func TestInterval_RoundTrip(t *testing.T) {
	original := Interval{Months: 1, Days: 15, Micros: 43200000000}

	val, _ := original.Value()
	witIv := val.(pg.Interval)
	fromRow := Interval{Months: witIv.Months, Days: witIv.Days, Micros: witIv.Micros}

	var recovered Interval
	require.NoError(t, recovered.Scan(fromRow))
	assert.Equal(t, original, recovered)
}

func TestInt32Range_Scan(t *testing.T) {
	tests := []struct {
		name    string
		src     any
		want    Int32Range
		wantErr bool
	}{{
		name: "nil src clears range",
		src:  nil,
		want: Int32Range{},
	}, {
		name: "both bounds inclusive",
		src: wittypes.Tuple2[
			wittypes.Option[wittypes.Tuple2[int32, pg.RangeBoundKind]],
			wittypes.Option[wittypes.Tuple2[int32, pg.RangeBoundKind]],
		]{
			F0: wittypes.Some(wittypes.Tuple2[int32, pg.RangeBoundKind]{F0: 1, F1: pg.RangeBoundKindInclusive}),
			F1: wittypes.Some(wittypes.Tuple2[int32, pg.RangeBoundKind]{F0: 10, F1: pg.RangeBoundKindInclusive}),
		},
		want: Int32Range{Lower: ptr(int32(1)), LowerInclusive: true, Upper: ptr(int32(10)), UpperInclusive: true},
	}, {
		name: "both bounds exclusive",
		src: wittypes.Tuple2[
			wittypes.Option[wittypes.Tuple2[int32, pg.RangeBoundKind]],
			wittypes.Option[wittypes.Tuple2[int32, pg.RangeBoundKind]],
		]{
			F0: wittypes.Some(wittypes.Tuple2[int32, pg.RangeBoundKind]{F0: 0, F1: pg.RangeBoundKindExclusive}),
			F1: wittypes.Some(wittypes.Tuple2[int32, pg.RangeBoundKind]{F0: 5, F1: pg.RangeBoundKindExclusive}),
		},
		want: Int32Range{Lower: ptr(int32(0)), LowerInclusive: false, Upper: ptr(int32(5)), UpperInclusive: false},
	}, {
		name: "unbounded lower",
		src: wittypes.Tuple2[
			wittypes.Option[wittypes.Tuple2[int32, pg.RangeBoundKind]],
			wittypes.Option[wittypes.Tuple2[int32, pg.RangeBoundKind]],
		]{
			F0: wittypes.None[wittypes.Tuple2[int32, pg.RangeBoundKind]](),
			F1: wittypes.Some(wittypes.Tuple2[int32, pg.RangeBoundKind]{F0: 100, F1: pg.RangeBoundKindExclusive}),
		},
		want: Int32Range{Lower: nil, LowerInclusive: false, Upper: ptr(int32(100)), UpperInclusive: false},
	}, {
		name: "unbounded upper",
		src: wittypes.Tuple2[
			wittypes.Option[wittypes.Tuple2[int32, pg.RangeBoundKind]],
			wittypes.Option[wittypes.Tuple2[int32, pg.RangeBoundKind]],
		]{
			F0: wittypes.Some(wittypes.Tuple2[int32, pg.RangeBoundKind]{F0: 5, F1: pg.RangeBoundKindInclusive}),
			F1: wittypes.None[wittypes.Tuple2[int32, pg.RangeBoundKind]](),
		},
		want: Int32Range{Lower: ptr(int32(5)), LowerInclusive: true, Upper: nil, UpperInclusive: false},
	}, {
		name:    "invalid src type",
		src:     "not a range",
		wantErr: true,
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var r Int32Range
			err := r.Scan(tt.src)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, r)
		})
	}
}

func TestInt32Range_Value(t *testing.T) {
	tests := []struct {
		name  string
		input Int32Range
		want  wittypes.Tuple2[
			wittypes.Option[wittypes.Tuple2[int32, pg.RangeBoundKind]],
			wittypes.Option[wittypes.Tuple2[int32, pg.RangeBoundKind]],
		]
	}{
		{
			name:  "both bounds inclusive",
			input: Int32Range{Lower: ptr(int32(1)), LowerInclusive: true, Upper: ptr(int32(10)), UpperInclusive: true},
			want: wittypes.Tuple2[
				wittypes.Option[wittypes.Tuple2[int32, pg.RangeBoundKind]],
				wittypes.Option[wittypes.Tuple2[int32, pg.RangeBoundKind]],
			]{
				F0: wittypes.Some(wittypes.Tuple2[int32, pg.RangeBoundKind]{F0: 1, F1: pg.RangeBoundKindInclusive}),
				F1: wittypes.Some(wittypes.Tuple2[int32, pg.RangeBoundKind]{F0: 10, F1: pg.RangeBoundKindInclusive}),
			},
		},
		{
			name:  "unbounded both",
			input: Int32Range{},
			want: wittypes.Tuple2[
				wittypes.Option[wittypes.Tuple2[int32, pg.RangeBoundKind]],
				wittypes.Option[wittypes.Tuple2[int32, pg.RangeBoundKind]],
			]{
				F0: wittypes.None[wittypes.Tuple2[int32, pg.RangeBoundKind]](),
				F1: wittypes.None[wittypes.Tuple2[int32, pg.RangeBoundKind]](),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, _ := tt.input.Value()
			assert.Equal(t, driver.Value(tt.want), got)
		})
	}
}

func TestInt32Range_RoundTrip(t *testing.T) {
	original := Int32Range{Lower: ptr(int32(3)), LowerInclusive: true, Upper: ptr(int32(7)), UpperInclusive: false}
	witVal, _ := original.Value()

	var recovered Int32Range
	recovered.Scan(witVal)
	assert.Equal(t, original, recovered)
}

func TestInt64Range_Scan(t *testing.T) {
	tests := []struct {
		name    string
		src     any
		want    Int64Range
		wantErr bool
	}{{
		name: "nil src clears range",
		src:  nil,
		want: Int64Range{},
	}, {
		name: "both bounds inclusive",
		src: wittypes.Tuple2[
			wittypes.Option[wittypes.Tuple2[int64, pg.RangeBoundKind]],
			wittypes.Option[wittypes.Tuple2[int64, pg.RangeBoundKind]],
		]{
			F0: wittypes.Some(wittypes.Tuple2[int64, pg.RangeBoundKind]{F0: 1, F1: pg.RangeBoundKindInclusive}),
			F1: wittypes.Some(wittypes.Tuple2[int64, pg.RangeBoundKind]{F0: 10, F1: pg.RangeBoundKindInclusive}),
		},
		want: Int64Range{Lower: ptr(int64(1)), LowerInclusive: true, Upper: ptr(int64(10)), UpperInclusive: true},
	}, {
		name: "both bounds exclusive",
		src: wittypes.Tuple2[
			wittypes.Option[wittypes.Tuple2[int64, pg.RangeBoundKind]],
			wittypes.Option[wittypes.Tuple2[int64, pg.RangeBoundKind]],
		]{
			F0: wittypes.Some(wittypes.Tuple2[int64, pg.RangeBoundKind]{F0: 0, F1: pg.RangeBoundKindExclusive}),
			F1: wittypes.Some(wittypes.Tuple2[int64, pg.RangeBoundKind]{F0: 5, F1: pg.RangeBoundKindExclusive}),
		},
		want: Int64Range{Lower: ptr(int64(0)), LowerInclusive: false, Upper: ptr(int64(5)), UpperInclusive: false},
	}, {
		name: "unbounded lower",
		src: wittypes.Tuple2[
			wittypes.Option[wittypes.Tuple2[int64, pg.RangeBoundKind]],
			wittypes.Option[wittypes.Tuple2[int64, pg.RangeBoundKind]],
		]{
			F0: wittypes.None[wittypes.Tuple2[int64, pg.RangeBoundKind]](),
			F1: wittypes.Some(wittypes.Tuple2[int64, pg.RangeBoundKind]{F0: 100, F1: pg.RangeBoundKindExclusive}),
		},
		want: Int64Range{Lower: nil, LowerInclusive: false, Upper: ptr(int64(100)), UpperInclusive: false},
	}, {
		name: "unbounded upper",
		src: wittypes.Tuple2[
			wittypes.Option[wittypes.Tuple2[int64, pg.RangeBoundKind]],
			wittypes.Option[wittypes.Tuple2[int64, pg.RangeBoundKind]],
		]{
			F0: wittypes.Some(wittypes.Tuple2[int64, pg.RangeBoundKind]{F0: 5, F1: pg.RangeBoundKindInclusive}),
			F1: wittypes.None[wittypes.Tuple2[int64, pg.RangeBoundKind]](),
		},
		want: Int64Range{Lower: ptr(int64(5)), LowerInclusive: true, Upper: nil, UpperInclusive: false},
	}, {
		name:    "invalid src type",
		src:     "not a range",
		wantErr: true,
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var r Int64Range
			err := r.Scan(tt.src)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, r)
		})
	}
}

func TestInt64Range_Value(t *testing.T) {
	tests := []struct {
		name  string
		input Int64Range
		want  wittypes.Tuple2[
			wittypes.Option[wittypes.Tuple2[int64, pg.RangeBoundKind]],
			wittypes.Option[wittypes.Tuple2[int64, pg.RangeBoundKind]],
		]
	}{{
		name:  "both bounds inclusive",
		input: Int64Range{Lower: ptr(int64(1)), LowerInclusive: true, Upper: ptr(int64(10)), UpperInclusive: true},
		want: wittypes.Tuple2[
			wittypes.Option[wittypes.Tuple2[int64, pg.RangeBoundKind]],
			wittypes.Option[wittypes.Tuple2[int64, pg.RangeBoundKind]],
		]{
			F0: wittypes.Some(wittypes.Tuple2[int64, pg.RangeBoundKind]{F0: 1, F1: pg.RangeBoundKindInclusive}),
			F1: wittypes.Some(wittypes.Tuple2[int64, pg.RangeBoundKind]{F0: 10, F1: pg.RangeBoundKindInclusive}),
		},
	}, {
		name:  "unbounded both",
		input: Int64Range{},
		want: wittypes.Tuple2[
			wittypes.Option[wittypes.Tuple2[int64, pg.RangeBoundKind]],
			wittypes.Option[wittypes.Tuple2[int64, pg.RangeBoundKind]],
		]{
			F0: wittypes.None[wittypes.Tuple2[int64, pg.RangeBoundKind]](),
			F1: wittypes.None[wittypes.Tuple2[int64, pg.RangeBoundKind]](),
		},
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, _ := tt.input.Value()
			assert.Equal(t, driver.Value(tt.want), got)
		})
	}
}

func TestInt64Range_RoundTrip(t *testing.T) {
	original := Int64Range{Lower: ptr(int64(3)), LowerInclusive: true, Upper: ptr(int64(7)), UpperInclusive: false}
	witVal, _ := original.Value()

	var recovered Int64Range
	recovered.Scan(witVal)
	assert.Equal(t, original, recovered)
}

func TestDecimal_Scan(t *testing.T) {
	tests := []struct {
		name    string
		src     any
		want    Decimal
		wantErr bool
	}{{
		name: "from string",
		src:  "123.456",
		want: Decimal("123.456"),
	}, {
		name: "large precision",
		src:  "99999999999999999999.99999999999999999999",
		want: Decimal("99999999999999999999.99999999999999999999"),
	}, {
		name: "nil src",
		src:  nil,
		want: Decimal(""),
	}, {
		name:    "invalid src type",
		src:     42,
		wantErr: true,
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var d Decimal
			err := d.Scan(tt.src)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, d)
		})
	}
}

func TestDecimal_Value(t *testing.T) {
	d := Decimal("123.456")
	got, _ := d.Value()
	assert.Equal(t, driver.Value(pg.MakeDbValueDecimal("123.456")), got)
}

func TestDecimalRange_Scan(t *testing.T) {
	tests := []struct {
		name    string
		src     any
		want    DecimalRange
		wantErr bool
	}{{
		name: "nil src clears range",
		src:  nil,
		want: DecimalRange{},
	}, {
		name: "both bounds inclusive",
		src: wittypes.Tuple2[
			wittypes.Option[wittypes.Tuple2[string, pg.RangeBoundKind]],
			wittypes.Option[wittypes.Tuple2[string, pg.RangeBoundKind]],
		]{
			F0: wittypes.Some(wittypes.Tuple2[string, pg.RangeBoundKind]{F0: "1.5", F1: pg.RangeBoundKindInclusive}),
			F1: wittypes.Some(wittypes.Tuple2[string, pg.RangeBoundKind]{F0: "9.99", F1: pg.RangeBoundKindInclusive}),
		},
		want: DecimalRange{Lower: ptr(Decimal("1.5")), LowerInclusive: true, Upper: ptr(Decimal("9.99")), UpperInclusive: true},
	}, {
		name: "unbounded lower",
		src: wittypes.Tuple2[
			wittypes.Option[wittypes.Tuple2[string, pg.RangeBoundKind]],
			wittypes.Option[wittypes.Tuple2[string, pg.RangeBoundKind]],
		]{
			F0: wittypes.None[wittypes.Tuple2[string, pg.RangeBoundKind]](),
			F1: wittypes.Some(wittypes.Tuple2[string, pg.RangeBoundKind]{F0: "100.00", F1: pg.RangeBoundKindExclusive}),
		},
		want: DecimalRange{Lower: nil, Upper: ptr(Decimal("100.00")), UpperInclusive: false},
	}, {
		name:    "invalid src type",
		src:     "not a range",
		wantErr: true,
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var r DecimalRange
			err := r.Scan(tt.src)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, r)
		})
	}
}

func TestDecimalRange_Value(t *testing.T) {
	tests := []struct {
		name  string
		input DecimalRange
		want  wittypes.Tuple2[
			wittypes.Option[wittypes.Tuple2[string, pg.RangeBoundKind]],
			wittypes.Option[wittypes.Tuple2[string, pg.RangeBoundKind]],
		]
	}{{
		name:  "both bounds inclusive",
		input: DecimalRange{Lower: ptr(Decimal("1.5")), LowerInclusive: true, Upper: ptr(Decimal("9.99")), UpperInclusive: true},
		want: wittypes.Tuple2[
			wittypes.Option[wittypes.Tuple2[string, pg.RangeBoundKind]],
			wittypes.Option[wittypes.Tuple2[string, pg.RangeBoundKind]],
		]{
			F0: wittypes.Some(wittypes.Tuple2[string, pg.RangeBoundKind]{F0: "1.5", F1: pg.RangeBoundKindInclusive}),
			F1: wittypes.Some(wittypes.Tuple2[string, pg.RangeBoundKind]{F0: "9.99", F1: pg.RangeBoundKindInclusive}),
		},
	}, {
		name:  "unbounded both",
		input: DecimalRange{},
		want: wittypes.Tuple2[
			wittypes.Option[wittypes.Tuple2[string, pg.RangeBoundKind]],
			wittypes.Option[wittypes.Tuple2[string, pg.RangeBoundKind]],
		]{
			F0: wittypes.None[wittypes.Tuple2[string, pg.RangeBoundKind]](),
			F1: wittypes.None[wittypes.Tuple2[string, pg.RangeBoundKind]](),
		},
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, _ := tt.input.Value()
			assert.Equal(t, driver.Value(tt.want), got)
		})
	}
}

func TestDecimalRange_RoundTrip(t *testing.T) {
	original := DecimalRange{Lower: ptr(Decimal("3.14")), LowerInclusive: true, Upper: ptr(Decimal("99.99")), UpperInclusive: false}
	witVal, _ := original.Value()

	var recovered DecimalRange
	recovered.Scan(witVal)
	assert.Equal(t, original, recovered)
}
