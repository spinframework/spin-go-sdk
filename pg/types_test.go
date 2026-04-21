package pg

import (
	"database/sql/driver"
	"testing"

	pg "github.com/spinframework/spin-go-sdk/v3/imports/spin_postgres_4_2_0_postgres"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	wittypes "go.bytecodealliance.org/pkg/wit/types"
)

func ptr[T any](v T) *T { return &v }

func TestInt32Range_Scan(t *testing.T) {
	tests := []struct {
		name    string
		src     any
		want    Int32Range
		wantErr bool
	}{
		{
			name: "nil src clears range",
			src:  nil,
			want: Int32Range{},
		},
		{
			name: "both bounds inclusive",
			src: wittypes.Tuple2[
				wittypes.Option[wittypes.Tuple2[int32, pg.RangeBoundKind]],
				wittypes.Option[wittypes.Tuple2[int32, pg.RangeBoundKind]],
			]{
				F0: wittypes.Some(wittypes.Tuple2[int32, pg.RangeBoundKind]{F0: 1, F1: pg.RangeBoundKindInclusive}),
				F1: wittypes.Some(wittypes.Tuple2[int32, pg.RangeBoundKind]{F0: 10, F1: pg.RangeBoundKindInclusive}),
			},
			want: Int32Range{Lower: ptr(int32(1)), LowerInclusive: true, Upper: ptr(int32(10)), UpperInclusive: true},
		},
		{
			name: "both bounds exclusive",
			src: wittypes.Tuple2[
				wittypes.Option[wittypes.Tuple2[int32, pg.RangeBoundKind]],
				wittypes.Option[wittypes.Tuple2[int32, pg.RangeBoundKind]],
			]{
				F0: wittypes.Some(wittypes.Tuple2[int32, pg.RangeBoundKind]{F0: 0, F1: pg.RangeBoundKindExclusive}),
				F1: wittypes.Some(wittypes.Tuple2[int32, pg.RangeBoundKind]{F0: 5, F1: pg.RangeBoundKindExclusive}),
			},
			want: Int32Range{Lower: ptr(int32(0)), LowerInclusive: false, Upper: ptr(int32(5)), UpperInclusive: false},
		},
		{
			name: "unbounded lower",
			src: wittypes.Tuple2[
				wittypes.Option[wittypes.Tuple2[int32, pg.RangeBoundKind]],
				wittypes.Option[wittypes.Tuple2[int32, pg.RangeBoundKind]],
			]{
				F0: wittypes.None[wittypes.Tuple2[int32, pg.RangeBoundKind]](),
				F1: wittypes.Some(wittypes.Tuple2[int32, pg.RangeBoundKind]{F0: 100, F1: pg.RangeBoundKindExclusive}),
			},
			want: Int32Range{Lower: nil, LowerInclusive: false, Upper: ptr(int32(100)), UpperInclusive: false},
		},
		{
			name: "unbounded upper",
			src: wittypes.Tuple2[
				wittypes.Option[wittypes.Tuple2[int32, pg.RangeBoundKind]],
				wittypes.Option[wittypes.Tuple2[int32, pg.RangeBoundKind]],
			]{
				F0: wittypes.Some(wittypes.Tuple2[int32, pg.RangeBoundKind]{F0: 5, F1: pg.RangeBoundKindInclusive}),
				F1: wittypes.None[wittypes.Tuple2[int32, pg.RangeBoundKind]](),
			},
			want: Int32Range{Lower: ptr(int32(5)), LowerInclusive: true, Upper: nil, UpperInclusive: false},
		},
		{
			name:    "invalid src type",
			src:     "not a range",
			wantErr: true,
		},
	}

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
			got, err := tt.input.Value()
			require.NoError(t, err)
			assert.Equal(t, driver.Value(tt.want), got)
		})
	}
}

func TestInt32Range_RoundTrip(t *testing.T) {
	original := Int32Range{Lower: ptr(int32(3)), LowerInclusive: true, Upper: ptr(int32(7)), UpperInclusive: false}

	witVal, err := original.Value()
	require.NoError(t, err)

	var recovered Int32Range
	require.NoError(t, recovered.Scan(witVal))
	assert.Equal(t, original, recovered)
}
