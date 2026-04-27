package util

import (
	"testing"

	sq "github.com/Masterminds/squirrel"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
)

func TestNormalizeOrderBy(t *testing.T) {
	t.Run("allow known columns and normalize direction", func(t *testing.T) {
		order, err := normalizeOrderBy(
			[]string{"name desc", "users.id"},
			[]string{"id", "name", "created_at"},
		)
		assert.NoError(t, err)
		assert.Equal(t, []string{"name DESC", "users.id ASC"}, order)
	})

	t.Run("allow empty order", func(t *testing.T) {
		order, err := normalizeOrderBy(nil, []string{"id"})
		assert.NoError(t, err)
		assert.Nil(t, order)
	})

	t.Run("reject unknown columns", func(t *testing.T) {
		_, err := normalizeOrderBy([]string{"email desc"}, []string{"id", "name"})
		assert.ErrorContains(t, err, "invalid order column")
	})

	t.Run("reject invalid directions", func(t *testing.T) {
		_, err := normalizeOrderBy([]string{"name sideways"}, []string{"id", "name"})
		assert.ErrorContains(t, err, "invalid order direction")
	})

	t.Run("reject overly complex clauses", func(t *testing.T) {
		_, err := normalizeOrderBy([]string{"name desc nulls last"}, []string{"id", "name"})
		assert.ErrorContains(t, err, "invalid order clause")
	})
}

func TestUpdateQRequiresWhere(t *testing.T) {
	err := UpdateQ(zerolog.Nop(), nil, "users", nil, map[string]any{"name": "alice"})
	assert.ErrorContains(t, err, "refuse update without where")

	err = UpdateQ(zerolog.Nop(), nil, "users", sq.Eq{"id": 1}, map[string]any{})
	assert.NoError(t, err)
}

func TestDeleteQRequiresWhere(t *testing.T) {
	err := DeleteQ(zerolog.Nop(), nil, "users", nil)
	assert.ErrorContains(t, err, "refuse delete without where")
}
