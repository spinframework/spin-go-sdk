package pg

import (
	"errors"
	"testing"

	pg "github.com/spinframework/spin-go-sdk/v3/imports/spin_postgres_4_2_0_postgres"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	wittypes "go.bytecodealliance.org/pkg/wit/types"
)

func TestToOptionSlice(t *testing.T) {
	got := toOptionSlice([]string{"a", "b", "c"})
	want := []wittypes.Option[string]{
		wittypes.Some("a"),
		wittypes.Some("b"),
		wittypes.Some("c"),
	}
	assert.Equal(t, want, got)
}

func TestFromOptionSlice(t *testing.T) {
	got := fromOptionSlice([]wittypes.Option[int]{
		wittypes.Some(1),
		wittypes.Some(2),
		wittypes.Some(3),
	})
	want := []int{1, 2, 3}
	assert.Equal(t, want, got)
}

func TestToError(t *testing.T) {
	t.Run("ConnectionFailed", func(t *testing.T) {
		err := toError(pg.MakeErrorConnectionFailed("connection refused"))
		assert.EqualError(t, err, "connection refused")
	})

	t.Run("BadParameter", func(t *testing.T) {
		err := toError(pg.MakeErrorBadParameter("invalid param"))
		assert.EqualError(t, err, "invalid param")
	})

	t.Run("QueryFailed/Text", func(t *testing.T) {
		err := toError(pg.MakeErrorQueryFailed(pg.MakeQueryErrorText("syntax error")))
		assert.EqualError(t, err, "syntax error")
	})

	t.Run("QueryFailed/DbError", func(t *testing.T) {
		dbErr := pg.DbError{
			AsText:   "ERROR 23505 (unique_violation): duplicate key",
			Severity: "ERROR",
			Code:     "23505",
			Message:  "duplicate key value violates unique constraint",
			Detail:   wittypes.Some("Key (id)=(1) already exists."),
			Extras:   []wittypes.Tuple2[string, string]{{F0: "constraint", F1: "users_pkey"}},
		}
		err := toError(pg.MakeErrorQueryFailed(pg.MakeQueryErrorDbError(dbErr)))
		require.Error(t, err)

		var pgErr *QueryDbError
		require.True(t, errors.As(err, &pgErr))
		assert.Equal(t, "ERROR", pgErr.Severity)
		assert.Equal(t, "23505", pgErr.Code)
		assert.Equal(t, "duplicate key value violates unique constraint", pgErr.Message)
		assert.Equal(t, "Key (id)=(1) already exists.", pgErr.Detail)
		assert.Equal(t, [][2]string{{"constraint", "users_pkey"}}, pgErr.Extras)
		assert.Contains(t, pgErr.Error(), "23505")
		assert.Contains(t, pgErr.Error(), "duplicate key value violates unique constraint")
		assert.Contains(t, pgErr.Error(), "Key (id)=(1) already exists.")
	})

	t.Run("QueryFailed/DbError/NoDetail", func(t *testing.T) {
		dbErr := pg.DbError{
			Severity: "ERROR",
			Code:     "42703",
			Message:  "column does not exist",
			Detail:   wittypes.None[string](),
		}
		err := toError(pg.MakeErrorQueryFailed(pg.MakeQueryErrorDbError(dbErr)))
		require.Error(t, err)

		var pgErr *QueryDbError
		require.True(t, errors.As(err, &pgErr))
		assert.Equal(t, "", pgErr.Detail)
		assert.Nil(t, pgErr.Extras)
	})

	t.Run("ValueConversionFailed", func(t *testing.T) {
		err := toError(pg.MakeErrorValueConversionFailed("cannot convert"))
		assert.EqualError(t, err, "cannot convert")
	})

	t.Run("Other", func(t *testing.T) {
		err := toError(pg.MakeErrorOther("unknown error"))
		assert.EqualError(t, err, "unknown error")
	})
}
