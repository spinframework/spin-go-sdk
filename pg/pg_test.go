package pg

import (
	"testing"

	"github.com/stretchr/testify/assert"
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
